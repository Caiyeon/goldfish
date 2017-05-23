package vault

import (
	"encoding/gob"
	"errors"
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
	// for authenticating this web server with vault
	vaultAddress  = ""
	vaultToken    = ""
	vaultClient   *api.Client

	// for initializing server and fetching client token from vault
	wrappingToken string
	rolePath      string
	configPath    string
	roleID        string
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})

	// setup flags to be parsed by server main()
	flag.StringVar(&vaultAddress, "vault_addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&wrappingToken, "vault_token", "", "The approle secret_id (must be in the form of a wrapping token)")
	flag.StringVar(&rolePath, "approle_path", "auth/approle/login", "The approle mount's login path")
	flag.StringVar(&configPath, "config_path", "", "A generic backend endpoint to store run-time settings. E.g. 'secret/goldfish'")
	flag.StringVar(&roleID, "role_id", "goldfish", "The approle role_id")
}

// fetches client token from vault using wrappingtoken and approle
func SetupServer(devMode bool) error {
	if vaultAddress == "" || wrappingToken == "" {
		return errors.New("Vault address and wrapped accessor missing. See --help for details")
	}
	if configPath == "" && !devMode {
		return errors.New("config_path must be set unless dev mode is enabled")
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

	// errors from go routines regarding server configuration is sent here
	errorChannel := make(chan error)

	if devMode && configPath == "" {
		// if devMode is active, unless configPath is set, load a set of simple configs
		loadDevModeConfig()
	} else {
		// load config once to ensure validity
		if err := loadConfigFromVault(configPath); err != nil {
			return err
		}
		go loadConfigEvery(time.Minute, errorChannel)
	}

	go renewServerTokenEvery(time.Hour, errorChannel)

	// to do: ship these potential error logs somewhere
	go func() {
		for err := range errorChannel {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()

	// report back the accessor so it may be safekept
	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
	return nil
}

func loadConfigEvery(interval time.Duration, ch chan error) {
	for {
		time.Sleep(interval)
		ch <- loadConfigFromVault(configPath)
	}
}

func renewServerTokenEvery(interval time.Duration, ch chan error) {
	for {
		time.Sleep(interval)
		ch <- renewServerToken()
	}
}
