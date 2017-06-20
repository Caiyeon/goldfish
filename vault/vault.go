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
	VaultAddress  = ""
	VaultSkipTLS  = false

	vaultToken    = ""
	vaultClient   *api.Client
	errorChannel  chan error
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})
}

func NewVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	err := config.ConfigureTLS(
		&api.TLSConfig{
			Insecure: VaultSkipTLS,
		},
	)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetAddress(VaultAddress)
	client.SetToken("")
	return client, nil
}

func StartGoldfishWrapper(wrappingToken, login, id string) error {
	if wrappingToken == "" {
		return errors.New("Token must be provided in non-dev mode")
	}

	client, err := NewVaultClient()
	if err != nil {
		return err
	}
	vaultClient = client

	// make a raw unwrap call. This will use the token as a header
	vaultClient.SetToken(wrappingToken)
	resp, err := vaultClient.Logical().Unwrap("")
	if err != nil {
		return errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}
	if resp == nil {
		return errors.New("Unwrap response from vault was nil. Please revoke token")
	}

	// verify that a secret_id was wrapped
	var secretID string
	err = errors.New("Could not find secret_id in wrapped token. Was it wrapped properly?")
	if raw, ok := resp.Data["secret_id"]; ok {
		if secretID, ok = raw.(string); ok {
			err = nil
		}
	}
	if err != nil {
		return err
	}

	// fetch vault token with secret_id
	resp, err = vaultClient.Logical().Write(login,
		map[string]interface{}{
			"role_id":   id,
			"secret_id": secretID,
		})
	if err != nil {
		return err
	}

	// verify that the secret_id is valid
	vaultToken = resp.Auth.ClientToken
	vaultClient.SetToken(resp.Auth.ClientToken)
	if _, err := vaultClient.Auth().Token().LookupSelf(); err != nil {
		return err
	}

	// errors that are not catastrophic can be logged here
	go func() {
		for err := range errorChannel {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()

	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
	return nil
}

func LoadRuntimeConfig(configPath string) error {
	// load config once to ensure validity
	if err := loadConfigFromVault(configPath); err != nil {
		return err
	}
	go loadConfigEvery(time.Minute, configPath)
	go renewServerTokenEvery(time.Hour)
	return nil
}

func loadConfigEvery(interval time.Duration, configPath string) {
	for {
		time.Sleep(interval)
		errorChannel <- loadConfigFromVault(configPath)
	}
}

func renewServerTokenEvery(interval time.Duration) {
	for {
		time.Sleep(interval)
		errorChannel <- renewServerToken()
	}
}
