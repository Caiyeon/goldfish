package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"io/ioutil"

	"github.com/caiyeon/goldfish/config"
	"github.com/caiyeon/goldfish/server"
	"github.com/caiyeon/goldfish/vault"
	"github.com/GeertJohan/go.rice"
	"github.com/hashicorp/vault/helper/mlock"
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
	staticAssets   *rice.Box
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
		fmt.Printf(devInitString)
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

	// start listener
	if !devMode {
		staticAssets, err = rice.FindBox("public")
		if err != nil {
			log.Fatalf("[ERROR]: Static assets not found. Build them with npm first. \n%s", err.Error())
		}
	}
	go server.StartListener(*cfg.Listener, staticAssets)
	fmt.Printf(versionString + initString)

	// wait for shutdown signal, and cleanup after
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	log.Println("\n\n==> Goldfish shutdown triggered")

	// shut down vault dev server, if it was initialized
	if devVaultCh != nil {
		close(devVaultCh)
	}

	// shut down listener, with a hard timeout
	server.StopListener(5 * time.Second)

	// extra grace time
	time.Sleep(time.Second)
	os.Exit(0)
}

const versionString = "Goldfish version: v0.9.1-dev"

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
                          See https://github.com/Caiyeon/goldfish/blob/master/config/sample.hcl
                          for a full list of options

Optional Arguments:

  -token=<uuid>           A wrapping token which contains a secret_id
                          Can be provided after launch, on Login page
                          Generate with 'vault write -f -wrap-ttl=5m auth/approle/role/goldfish/secret-id'

  -nomad-token-file       A path to a file containing a raw token.
                          Not recommended unless approle is unavailable,
						  in the case of Nomad for example.

  -version                Print the version and exit

  -dev                    Launch goldfish in dev mode
                          A localhost dev vault instance will be launched
`
