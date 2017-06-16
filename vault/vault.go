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
	vaultToken    = ""

	vaultClient   *api.Client
	vaultSkipTLS  = false

	configPath    = ""
	approleID     = ""
	approleLogin  = ""
)

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})
}

func NewVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	err := config.ConfigureTLS(
		&api.TLSConfig{
			Insecure: vaultSkipTLS,
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

func StartGoldfishWrapper(wrappingToken string) error {
	if wrappingToken == "" {
		return errors.New("vault_token cannot be empty")
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

	// verify that a secret_id was wrapped
	secretID, ok := resp.Data["secret_id"].(string)
	if !ok {
		return errors.New("Failed to unwrap provided token, revoke it if possible")
	}

	// fetch vault token with secret_id
	resp, err = vaultClient.Logical().Write(approleLogin,
		map[string]interface{}{
			"role_id":   approleID,
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

func ParseDeploymentConfig(cfg map[string]string) error {
	skip, ok := cfg["tls_skip_verify"]
	if ok && skip != "0" {
		if skip == "1" {
			vaultSkipTLS = true
		} else {
			return errors.New("Config: vault.tls_skip_verify can be '1' or '0'")
		}
	}

	VaultAddress, ok = cfg["address"]
	if !ok {
		return errors.New("Config: vault.address must be set")
	}

	configPath, ok = cfg["runtime_config"]
	if !ok {
		return errors.New("Config: vault.runtime_config must be set")
	}

	approleID, ok = cfg["approle_id"]
	if !ok {
		return errors.New("Config: vault.approle_id must be set")
	}

	approleLogin, ok = cfg["approle_login"]
	if !ok {
		return errors.New("Config: vault.approle_login must be set")
	}

	return nil
}

func LoadRuntimeConfig(devMode bool, errorChannel chan error) error {
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
