package vault

import (
	"encoding/gob"
	"flag"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

type AuthInfo struct {
	Type string `json:"Type" form:"Type" query:"Type"`
	ID   string `json:"ID" form:"ID" query:"ID"`
	Pass string `json:"password" form:"Password" query:"Password"`
}

var (
	// If "-dev" is provided, many safe defaults are turned off
	DevMode      = false

	// for authenticating this web server with vault
	vaultAddress = ""
	vaultToken   = ""
	vaultClient  *api.Client
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})

	flag.BoolVar(&DevMode, "dev", false, "Set to true to save time in development. DO NOT SET TO TRUE IN PRODUCTION!!")

	// read address and wrapping token inputs
	var wrappingToken, roleID, rolePath, configPath string
	flag.StringVar(&vaultAddress, "vault_addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&wrappingToken, "vault_token", "", "The approle secret_id (must be in the form of a wrapping token)")
	flag.StringVar(&rolePath, "approle_path", "auth/approle/login", "The approle mount's login path")
	flag.StringVar(&configPath, "config_path", "secret/goldfish", "A generic backend endpoint to store run-time settings")
	flag.StringVar(&roleID, "role_id", "goldfish", "The approle role_id")

	flag.Parse()
	if vaultAddress == "" || wrappingToken == "" {
		panic("Invalid cmd args. See --help for details")
	}

	// setup server's vault client, used for login transit encryption/decryption
	resp, err := loginWithSecretID(vaultAddress, wrappingToken, roleID, rolePath)
	if err != nil {
		panic(err)
	}
	vaultToken = resp.Auth.ClientToken
	vaultClient, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	vaultClient.SetAddress(vaultAddress)
	vaultClient.SetToken(vaultToken)

	if DevMode {
		loadDevModeConfig()
	} else {
		// load config once to ensure validity
		if err := loadConfigFromVault(configPath); err != nil {
			panic(err)
		} else {
			// continuously load config in go routine
			go func() {
				for {
					time.Sleep(time.Minute)
					if err := loadConfigFromVault(configPath); err != nil {
						log.Println(err)
					} // if there are errors, just try again in a minute
				}
			}()
		}
	}

	// report back the accessor so it may be safekept
	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
}
