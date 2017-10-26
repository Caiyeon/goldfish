package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"io/ioutil"

	"github.com/caiyeon/goldfish/config"
	"github.com/caiyeon/goldfish/handlers"
	"github.com/caiyeon/goldfish/vault"
	"github.com/hashicorp/vault/helper/mlock"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	rice "github.com/GeertJohan/go.rice"

	"golang.org/x/crypto/acme/autocert"
)

var (
	cfg            *config.Config
	cfgPath        string
	devMode        bool
	devVaultCh     chan struct{}
	err            error
	nomadTokenFile string
	printVersion   bool
	wrappingToken  string
)

func init() {
	// customized help message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, helpMessage)
	}

	// cmd line args
	flag.BoolVar(&devMode, "dev", false, "Set to true to save time in development. DO NOT SET TO TRUE IN PRODUCTION!!")
	flag.BoolVar(&printVersion, "version", false, "Display goldfish's version and exit")
	flag.StringVar(&wrappingToken, "token", "", "Token generated from approle (must be wrapped!)")
	flag.StringVar(&nomadTokenFile, "nomad-token-file", "", "If you are using Nomad, this file should contain a secret_id")
	flag.StringVar(&cfgPath, "config", "", "The path of the deployment config HCL file")

	// if vault dev core is active, relay shutdown signal
	shutdownCh := make(chan os.Signal, 4)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdownCh
		log.Println("\n\n==> Goldfish shutdown triggered")
		if devVaultCh != nil {
			close(devVaultCh)
		}
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}

func main() {
	// if --version, print and exit success
	flag.Parse()
	if printVersion {
		log.Println(versionString)
		os.Exit(0)
	}

	// if dev mode, run a localhost dev vault instance
	if devMode {
		var unsealTokens []string
		cfg, devVaultCh, unsealTokens, wrappingToken, err = config.LoadConfigDev()
		log.Println("[INFO ]: Dev mode wrapping token: " + wrappingToken)
		log.Println("[INFO ]: Dev mode unseal tokens:\n" + strings.Join(unsealTokens, "\n"))
	} else {
		cfg, err = config.LoadConfigFile(cfgPath)
	}
	if err != nil {
		log.Fatalf("[ERROR]: Launching goldfish: %s", err.Error())
	}

	if !cfg.DisableMlock {
		if err := mlock.LockMemory(); err != nil {
			log.Fatalf(mlockError, err.Error())
		}
	}

	// configure goldfish server settings and token
	vault.SetConfig(cfg.Vault)

	// if bootstrapping options are provided, do so immediately
	if wrappingToken != "" {
		if err := vault.Bootstrap(wrappingToken); err != nil {
			log.Fatalf("[ERROR]: Bootstrapping goldfish %s", err.Error())
		}
	} else if nomadTokenFile != "" {
		raw, err := ioutil.ReadFile(nomadTokenFile)
		if err != nil {
			log.Fatalf("[ERROR]: Could not read token file: %s", err.Error())
		}
		if err := vault.BootstrapRaw(string(raw)); err != nil {
			log.Fatalf("[ERROR]: Bootstrapping goldfish: %s", err.Error())
		}
	}


	// display welcome message
	if devMode {
		fmt.Printf(devInitString)
	}
	fmt.Printf(versionString + initString)

	// instantiate echo web server
	e := echo.New()
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

	// unless explicitly disabled, some extra https configurations need to be set
	if !cfg.Listener.Tls_disable {
		// add extra security headers
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self' https://api.github.com/repos/caiyeon/goldfish",
		}))

		// if redirect is set, forward port 80 to port 443
		if cfg.Listener.Tls_autoredirect {
			e.Pre(middleware.HTTPSRedirect())
			go func(c *echo.Echo) {
				e.Logger.Fatal(e.Start(":80"))
			}(e)
		}

		// if cert file and key file are not provided, try using let's encrypt
		if cfg.Listener.Tls_cert_file == "" && cfg.Listener.Tls_key_file == "" {
			e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(cfg.Listener.Address)
			e.Use(middleware.HTTPSRedirectWithConfig(middleware.RedirectConfig{
				Code: 301,
			}))
		}
	}

	// for production, static files are packed inside binary
	// for development, npm dev should serve the static files instead
	if !devMode {
		// use rice for static files instead of regular file system
		assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
		e.GET("/", echo.WrapHandler(assetHandler))
		e.GET("/assets/css/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/js/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/fonts/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/assets/img/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	}

	// API routing
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

	// serving both static folder and API
	if cfg.Listener.Tls_disable {
		// launch http-only listener
		e.Logger.Fatal(e.Start(cfg.Listener.Address))
	} else if cfg.Listener.Tls_cert_file == "" && cfg.Listener.Tls_key_file == "" {
		// if https is enabled, but no cert provided, try let's encrypt
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		// launch listener in https
		e.Logger.Fatal(e.StartTLS(
			cfg.Listener.Address,
			cfg.Listener.Tls_cert_file,
			cfg.Listener.Tls_key_file,
		))
	}
}

const versionString = "Goldfish version: v0.7.5-custom"

const devInitString = `

---------------------------------------------------
Starting local vault dev instance...
Your unseal token and root token can be found above
`

const initString = `
Goldfish successfully bootstrapped to vault

  .
  ...             ...
  .........       ......
   ...........   ..........
     .......... ...............
     .............................
      .............................
         ...........................
        ...........................
        ..........................
        ...... ..................
      ......    ...............
     ..        ..      ....
    .                 ..


`

const mlockError = `
Failed to use mlock to prevent swap usage: %s

Goldfish uses mlock similar to Vault. See here for details:
https://www.vaultproject.io/docs/configuration/index.html#disable_mlock

To enable mlock without launching goldfish as root:
sudo setcap cap_ipc_lock=+ep $(readlink -f $(which goldfish))

To disable mlock entirely, set disable_mlock to "1" in config file
`

const helpMessage = `Usage: goldfish [options]
See https://github.com/Caiyeon/goldfish/wiki for details

Required Arguments:

  -config=config.hcl      The deployment config file
                          See github.com/caiyeon/goldfish/config/sample.hcl
                          for a full list of options

Optional Arguments:

  -token=<uuid>           A wrapping token which contains a secret_id
                          Can be provided after launch, on Login page
                          Generate with 'vault write -f transit/keys/goldfish'

  -nomad-token-file       A path to a file containing a raw token.
                          Not recommended unless approle is unavailable,
						  in the case of Nomad for example.

  -version                Print the version and exit

  -dev                    Launch goldfish in dev mode
                          A localhost dev vault instance will be launched
`
