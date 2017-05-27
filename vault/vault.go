package vault

import (
	"encoding/gob"
	"errors"
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
	configPath    string
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})
}

func SetAddress(addr string) error {
	config := api.DefaultConfig()
	err := config.ConfigureTLS(
		&api.TLSConfig{
			// in preparation for tls_skip_verify option
			Insecure: false,
		},
	)
	if err != nil {
		return err
	}
	config.Address = addr
	vaultAddress = addr

	vaultClient, err = api.NewClient(config)
	if err != nil {
		return err
	}
	return nil
}

func UnwrapSecretID(wrappingToken, roleID, rolePath string) error {
	// make a raw unwrap call. This will use the token as a header
	vaultClient.SetToken(wrappingToken)
	resp, err := vaultClient.Logical().Unwrap("")
	if err != nil {
		errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}

	// verify that a secret_id was wrapped
	secretID, ok := resp.Data["secret_id"].(string)
	if !ok {
		errors.New("Failed to unwrap provided token, revoke it if possible")
	}

	// fetch vault token with secret_id
	resp, err = vaultClient.Logical().Write(rolePath,
		map[string]interface{}{
			"role_id":   roleID,
			"secret_id": secretID,
		})
	if err != nil {
		return err
	}

	// verify that the secret_id is valid
	vaultToken = resp.Auth.ClientToken
	vaultClient.SetToken(resp.Auth.ClientToken)
	if _, err = vaultClient.Auth().Token().LookupSelf(); err != nil {
		return err
	}

	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
	return nil
}

func LoadConfig(devMode bool, config string, errorChannel chan error) error {
	configPath = config
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

func loginWithSecretID(address, token, roleID, rolePath string) (*api.Secret, error) {
	// set up vault client
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(address)
	client.SetToken(token)

	// make a raw unwrap call. This will use the token as a header
	resp, err := client.Logical().Unwrap("")
	if err != nil {
		return nil, errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}

	// verify that a secret_id was wrapped
	secretID, ok := resp.Data["secret_id"].(string)
	if !ok {
		return nil, errors.New("Failed to unwrap provided token, revoke it if possible")
	}

	// fetch vault token with secret_id
	resp, err = client.Logical().Write(rolePath,
		map[string]interface{}{
			"role_id":   roleID,
			"secret_id": secretID,
		})
	if err != nil {
		return nil, err
	}

	// verify that the secret_id is valid
	client.SetToken(resp.Auth.ClientToken)
	_, err = client.Auth().Token().LookupSelf()
	return resp, err
}