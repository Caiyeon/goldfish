package main

import (
	"flag"
	"log"
	"os"

	"github.com/caiyeon/goldfish/handlers"
	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"golang.org/x/crypto/acme/autocert"
)

const versionString = "Goldfish version: v0.3.3"

var (
	devMode         bool
	goldfishAddress string
	certFile        string
	keyFile         string
	wrappingToken   string
	rolePath        string
	roleID          string
	printVersion    bool
	tlsDisable      bool
)

func init() {
	// dev mode flag. This will disable a lot of HTTPS defaults
	flag.BoolVar(&devMode, "dev", false, "Set to true to save time in development. DO NOT SET TO TRUE IN PRODUCTION!!")

	// web server config
	flag.StringVar(&goldfishAddress, "goldfish_addr", "http://127.0.0.1:8000", "Goldfish server's listening address")
	flag.StringVar(&certFile, "cert_file", "", "Goldfish server's certificate")
	flag.StringVar(&keyFile, "key_file", "", "Goldfish certificate's private key file")
	flag.BoolVar(&tlsDisable, "tls_disable", false, "Disables HTTPS. Be careful when turning this option on!")

	// vault wrapper package config
	flag.BoolVar(&vault.VaultSkipTLS, "tls_skip_verify", false, "Set to true to not verify vault's certificate (e.g. if it was self-signed")
	flag.StringVar(&vault.VaultAddress, "vault_addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&vault.ConfigPath, "config_path", "", "A generic backend endpoint to store run-time settings. E.g. 'secret/goldfish'")

	// vault wrapper login info
	flag.StringVar(&rolePath, "approle_path", "auth/approle/login", "The approle mount's login path")
	flag.StringVar(&roleID, "role_id", "goldfish", "The approle role_id")
	flag.StringVar(&wrappingToken, "vault_token", "", "The approle secret_id (must be in the form of a wrapping token)")

	// if true, print version and exit
	flag.BoolVar(&printVersion, "version", false, "Display goldfish's version and exit")
}

func main() {
	// parse cmd args
	flag.Parse()

	// if --version, print and exit success
	if (printVersion) {
		log.Println(versionString)
		os.Exit(0)
	}

	// non-fatal wrapper errors should be sent here and logged
	errorChannel := make(chan error)
	go func() {
		for err := range errorChannel {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()

	// if API wrapper can't start, panic is justified
	if err := vault.StartGoldfishWrapper(wrappingToken, roleID, rolePath); err != nil {
		panic(err)
	}
	// load config from vault, and start goroutines for token renewal & config hot reload
	if err := vault.LoadConfig(devMode, errorChannel); err != nil {
		panic(err)
	}

	// instantiate echo web server
	e := echo.New()

	// setup middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echo.WrapMiddleware(
		csrf.Protect(
			// Generate a new encryption key for cookies each launch
			// invalidating previous goldfish instance's cookies is purposeful
			[]byte(securecookie.GenerateRandomKey(32)),
			// when devMode is false, cookie will only be sent through https
			csrf.Secure(!devMode || tlsDisable),
		)))

	// add security headers if deployment is production
	if (!devMode || tlsDisable) {
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self'",
		}))
	}

	// if cert and key are not provided, try using let's encrypt
	if !devMode && certFile == "" && keyFile == "" {
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(goldfishAddress)
		e.Use(middleware.HTTPSRedirectWithConfig(middleware.RedirectConfig{
			Code: 301,
		}))
	}

	// static routing of webpack'd folder
	e.Static("/", "public")

	// API routing
	e.GET("/api/health", handlers.VaultHealth())

	e.GET("/api/login/csrf", handlers.FetchCSRF())
	e.POST("/api/login", handlers.Login())
	e.POST("/api/login/renew-self", handlers.RenewSelf())

	e.GET("/api/users", handlers.GetUsers())
	e.GET("/api/users/csrf", handlers.FetchCSRF())
	e.GET("/api/tokencount", handlers.GetTokenCount())
	e.GET("/api/users/role", handlers.GetRole())
	e.GET("/api/users/listroles", handlers.ListRoles())
	e.POST("/api/users/revoke", handlers.DeleteUser())
	e.POST("/api/users/create", handlers.CreateUser())

	e.GET("/api/policy", handlers.GetPolicy())
	e.DELETE("/api/policy", handlers.DeletePolicy())

	e.GET("/api/policy/request", handlers.GetPolicyRequest())
	e.POST("/api/policy/request", handlers.AddPolicyRequest())
	e.POST("/api/policy/request/update", handlers.UpdatePolicyRequest())
	e.DELETE("/api/policy/request/:id", handlers.DeletePolicyRequest())

	e.GET("/api/transit", handlers.TransitInfo())
	e.POST("/api/transit/encrypt", handlers.EncryptString())
	e.POST("/api/transit/decrypt", handlers.DecryptString())

	e.GET("/api/mounts", handlers.GetMounts())
	e.GET("/api/mounts/:mountname", handlers.GetMount())
	e.POST("/api/mounts/:mountname", handlers.ConfigMount())

	e.GET("/api/secrets", handlers.GetSecrets())
	e.POST("/api/secrets", handlers.PostSecrets())

	e.GET("/api/bulletins", handlers.GetBulletins())

	e.GET("/api/wrapping", handlers.FetchCSRF())
	e.POST("/api/wrapping/wrap", handlers.WrapHandler())
	e.POST("/api/wrapping/unwrap", handlers.UnwrapHandler())

	// serving both static folder and API
	if (devMode) {
		// start the server in HTTP. DO NOT USE THIS IN PRODUCTION!!
		e.Logger.Fatal(e.Start("127.0.0.1:8000"))
	} else if (tlsDisable) {
		// if https is disabled, listen at given address
		e.Logger.Fatal(e.Start(goldfishAddress))
	} else if certFile == "" && keyFile == "" {
		// if cert and key file arent provided, try let's encrypt
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		e.Logger.Fatal(e.StartTLS(goldfishAddress, certFile, keyFile))
	}
}
