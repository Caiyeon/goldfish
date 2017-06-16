package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"os"
	"os/signal"
	"syscall"

	"github.com/caiyeon/goldfish/config"
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
	devMode       bool
	wrappingToken string
	cfgPath       string
	cfg           *config.Config
	devVaultCh    chan struct{}
	err           error
	printVersion  bool
)

func init() {
	flag.BoolVar(&devMode, "dev", false, "Set to true to save time in development. DO NOT SET TO TRUE IN PRODUCTION!!")
	flag.BoolVar(&printVersion, "version", false, "Display goldfish's version and exit")
	flag.StringVar(&wrappingToken, "vault_token", "", "The approle secret_id (must be in the form of a wrapping token)")
	flag.StringVar(&cfgPath, "config", "", "The path of the deployment config HCL file")
}

func main() {
	// parse cmd args
	flag.Parse()

	// if --version, print and exit success
	if printVersion {
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

	// if dev mode, run a localhost dev vault instance
	if devMode {
		cfg, devVaultCh, wrappingToken, err = config.LoadConfigDev()
	} else {
		cfg, err = config.LoadConfigFile(cfgPath)
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf(initString)

	// if API wrapper can't start, panic is justified
	if err := vault.ParseDeploymentConfig(cfg.Vault.Config); err != nil {
		panic(err)
	}
	if err := vault.StartGoldfishWrapper(wrappingToken); err != nil {
		panic(err)
	}
	if err := vault.LoadRuntimeConfig(devMode, errorChannel); err != nil {
		panic(err)
	}

	// if dev or tls_disable, set https to false
	https := !devMode
	if tlsDisable, ok := cfg.Listener.Config["tls_disable"]; ok && tlsDisable != "0" {
		if tlsDisable == "1" {
			https = false
		} else {
			panic("Config: listener.tls_disable can be either '0' or '1'")
		}
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
			csrf.Secure(https),
		)))

	// add security headers if deployment is in https mode
	if https {
		e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self'",
		}))
	}

	goldfishAddress, ok := cfg.Listener.Config["address"]
	if !ok {
		if devMode {
			goldfishAddress = "http://127.0.0.1:8000"
		} else {
			panic("Config: listener.Address cannot be empty")
		}
	}
	certFile, ok := cfg.Listener.Config["tls_cert_file"]
	if !ok {
		certFile = ""
	}
	keyFile, ok := cfg.Listener.Config["tls_key_file"]
	if !ok {
		keyFile = ""
	}

	// if cert and key are not provided, try using let's encrypt
	if https && certFile == "" && keyFile == "" {
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

	// upon sigint, give vault dev core time to shut down
	if (devMode) {
		shutdownCh := make(chan os.Signal, 4)
		signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<- shutdownCh
			close(devVaultCh)
			time.Sleep(time.Second)
			os.Exit(0)
		}()
	}

	// serving both static folder and API
	if (devMode) {
		// start the server in HTTP. DO NOT USE THIS IN PRODUCTION!!
		e.Logger.Fatal(e.Start("127.0.0.1:8000"))
	} else if (!https) {
		// if https is disabled, listen at given address
		e.Logger.Fatal(e.Start(goldfishAddress))
	} else if certFile == "" && keyFile == "" {
		// if cert and key file arent provided, try let's encrypt
		e.Logger.Fatal(e.StartAutoTLS(":443"))
	} else {
		e.Logger.Fatal(e.StartTLS(
			goldfishAddress,
			certFile,
			keyFile,
		))
	}
}

const initString = `


---------------------------------------------------
Starting local vault dev instance...
Your unseal token and root token can be found above

`
