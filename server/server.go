package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"time"
	"log"
	"strings"
	"sync"

	"github.com/caiyeon/goldfish/config"
	"github.com/caiyeon/goldfish/handlers"
	"github.com/caiyeon/goldfish/vault"
	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/crypto/acme/autocert"
)

var (
	e        *echo.Echo
	cert     *tls.Certificate
	certLock = new(sync.RWMutex)
)

func StartListener(listener config.ListenerConfig, assets *rice.Box) {
	// already configured, restarting listener at runtime is not currently supported
	if e != nil {
		return
	}

	// instantiate echo instance
	e = echo.New()
	e.HideBanner = true
	e.Server.ReadTimeout = 10 * time.Second
	e.Server.WriteTimeout = 2 * time.Minute

	// setup middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("32M"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// prevent caching by client (e.g. Safari)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			return next(c)
		}
	})

	// by default, some security features will accompany https listeners
	if !listener.Tls_disable {
		// add extra security headers
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self' https://api.github.com/repos/caiyeon/goldfish",
		}))

		// if auto-redirect is set, forward port 80 to port 443
		if listener.Tls_autoredirect {
			e.Pre(middleware.HTTPSRedirect())
			go func(c *echo.Echo) {
				e.Logger.Fatal(e.Start(":80"))
			}(e)
		}

		// if cert file and key file are not provided, try using let's encrypt
		if listener.Tls_cert_file == "" && listener.Tls_key_file == "" {
			e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(listener.Address)
			e.Use(middleware.HTTPSRedirectWithConfig(middleware.RedirectConfig{
				Code: 301,
			}))
		}
	}

	// if this is production, static files must be already packed
	// if they don't exist, exit with error
	if assets != nil {
		assetHandler := http.FileServer(assets.HTTPBox())
		e.GET("/", echo.WrapHandler(assetHandler))
		e.GET("/assets/css/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/js/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/fonts/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/img/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	}

	// setup API routes
	e.GET("/v1/health", handlers.Health())
	e.GET("/v1/vaulthealth", handlers.VaultHealth())
	e.POST("/v1/bootstrap", handlers.Bootstrap())

	e.POST("/v1/login", handlers.Login())
	e.POST("/v1/login/renew-self", handlers.RenewSelf())

	e.GET("/v1/token/accessors", handlers.GetTokenAccessors())
	e.POST("/v1/token/lookup-accessor", handlers.LookupTokenByAccessor())
	e.POST("/v1/token/revoke-accessor", handlers.RevokeTokenByAccessor())
	e.POST("/v1/token/create", handlers.CreateToken())
	e.GET("/v1/token/listroles", handlers.ListRoles())
	e.GET("/v1/token/role", handlers.GetRole())

	e.GET("/v1/userpass/users", handlers.GetUserpassUsers())
	e.POST("/v1/userpass/delete", handlers.DeleteUserpassUser())

	e.GET("/v1/approle/roles", handlers.GetApproleRoles())
	e.POST("/v1/approle/delete", handlers.DeleteApproleRole())

	e.GET("/v1/ldap/groups", handlers.GetLDAPGroups())
	e.GET("/v1/ldap/users", handlers.GetLDAPUsers())

	e.GET("/v1/policy", handlers.GetPolicy())
	e.DELETE("/v1/policy", handlers.DeletePolicy())

	e.GET("/v1/request", handlers.GetRequest())
	e.POST("/v1/request/add", handlers.AddRequest())
	e.POST("/v1/request/approve", handlers.ApproveRequest())
	e.DELETE("/v1/request/reject", handlers.RejectRequest())

	e.GET("/v1/transit", handlers.TransitInfo())
	e.POST("/v1/transit/encrypt", handlers.EncryptString())
	e.POST("/v1/transit/decrypt", handlers.DecryptString())

	e.GET("/v1/mount", handlers.GetMount())
	e.POST("/v1/mount", handlers.ConfigMount())

	e.GET("/v1/secrets", handlers.GetSecrets())
	e.POST("/v1/secrets", handlers.PostSecrets())
	e.DELETE("/v1/secrets", handlers.DeleteSecrets())

	e.GET("/v1/bulletins", handlers.GetBulletins())

	e.POST("/v1/wrapping/wrap", handlers.WrapHandler())
	e.POST("/v1/wrapping/unwrap", handlers.UnwrapHandler())

	// start listening on configured port
	// launch http-only listener
	if listener.Tls_disable {
		e.Logger.Fatal(e.Start(listener.Address))

	// fetch certificate from vault PKI backend
	} else if listener.Tls_PKI_path != "" {
		c, err := vault.FetchCertificate(
			listener.Tls_PKI_path,
			// hardcoded due to vagrant limitations
			strings.Split(listener.Address, ":")[0],
		)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
		certLock.Lock()
		cert = c
		certLock.Unlock()

		// start background job to monitor certificate expiry and periodically renew
		go maintainCertificate(listener.Tls_PKI_path, strings.Split(listener.Address, ":")[0])

		// start custom echo server with getcertificate function
		e.TLSServer.TLSConfig = new(tls.Config)
		e.TLSServer.TLSConfig.GetCertificate = GetCertificate
		e.TLSServer.Addr = listener.Address
		e.Logger.Fatal(e.StartServer(e.TLSServer))

	// if https is enabled, but no cert provided, try let's encrypt
	} else if listener.Tls_cert_file == "" && listener.Tls_key_file == "" {
		e.Logger.Fatal(e.StartAutoTLS(":443"))

	// launch listener in https with certificate from files on local system
	} else {
		e.Logger.Fatal(e.StartTLS(
			listener.Address,
			listener.Tls_cert_file,
			listener.Tls_key_file,
		))
	}
}

func StopListener(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certLock.RLock()
	defer certLock.RUnlock()

	if cert == nil {
		return nil, errors.New("No certificate configured.")
	}

	return cert, nil
}

func maintainCertificate(path, url string) {
	// check the certificate's expiry date
	certLock.RLock()
	if cert == nil || len(cert.Certificate) == 0 {
		return
	}
	x509c, err := x509.ParseCertificate(cert.Certificate[0])
	certLock.RUnlock()

	if err != nil {
		return
	}
	notafter := x509c.NotAfter

	// loop forever
	for {
		// sleep till halfway to expiry date
		time.Sleep(notafter.Sub(time.Now())/2)

		// fetch new certificate from vault
		for {
			if c, err := vault.FetchCertificate(path, url); err != nil {
				log.Println("[ERROR]: Error fetching certificate from PKI backend", err.Error())

			} else if len(c.Certificate) > 0 {
				// recalculate next interval
				x509c, err = x509.ParseCertificate(cert.Certificate[0])
				if err == nil {
					notafter = x509c.NotAfter

					// replace certificate
					certLock.Lock()
					cert = c
					certLock.Unlock()
					log.Println("[INFO ]: Certificate replaced from PKI backend")

					// break inner loop on success
					break
				}
			}

			// re-try in 30 minutes if failed to get certificate
			time.Sleep(30 * time.Minute)
		}
	}
}
