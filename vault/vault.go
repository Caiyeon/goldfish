package vault

import (
	"encoding/gob"
	"flag"
	"log"
	"time"

	"github.com/hashicorp/vault/api"
)

type AuthInfo struct {
	Type     string `json:"Type" form:"Type" query:"Type"`
	ID       string `json:"ID" form:"ID" query:"ID"`
	Pass     string `json:"password" form:"Password" query:"Password"`
}

// for authenticating this web server with vault
var (
	vaultAddress = ""
	vaultToken = ""
	vaultClient *api.Client
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})

	// read address and wrapping token inputs
	var wrappingToken, roleID, rolePath, configPath string
	flag.StringVar(&vaultAddress, "addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&wrappingToken, "token", "", "Wrapping token that should contain a secret_id")
	flag.StringVar(&roleID, "role_id", "goldfish", "The role_id the secret_id was generated from")
	flag.StringVar(&rolePath, "approle_path", "auth/approle/login", "The login path of the mount e.g. 'auth/approle/login'")
	flag.StringVar(&configPath, "config_path", "data/goldfish", "The vault path containing goldfish config data e.g. 'secret/goldfish'")
	flag.Parse()
	if vaultAddress == "" || wrappingToken == "" {
		panic("Provide credentials via -addr and -token (wrapping token only)")
	}

	// setup server's vault client, used for login transit encryption/decryption
	resp, err := fetchAppRole(vaultAddress, wrappingToken, roleID, rolePath)
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

	// load config once to ensure validity
	if err := loadConfigFromVault(configPath); err != nil {
		panic(err)
	} else {
		// continuously load config in go routine
		go func() {
			time.Sleep(5 * time.Second)
			if err := loadConfigFromVault(configPath); err != nil {
				log.Println(err)
			}
		}()
	}

	// report back the accessor so it may be safekept
	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
}
