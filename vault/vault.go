package vault

import (
	"errors"
	"log"
	"time"

	"github.com/caiyeon/goldfish/config"
	"github.com/hashicorp/vault/api"
)

type AuthInfo struct {
	Type string `json:"Type" form:"Type" query:"Type"`
	ID   string `json:"ID" form:"ID" query:"ID"`
	Pass string `json:"password" form:"Password" query:"Password"`
}

var (
	vaultConfig  config.VaultConfig
	vaultToken   string
	errorChannel = make(chan error, 1)
)

func init() {
	// non-catastrophic errors can be logged via errorChannel
	// e.g. if goldfish server was unable to fetch runtime config
	go func() {
		for err := range errorChannel {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()
}

func Bootstrapped() bool {
	return vaultToken != ""
}

func SetConfig(c *config.VaultConfig) {
	vaultConfig = *c
}

func NewVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	err := config.ConfigureTLS(
		&api.TLSConfig{
			Insecure: vaultConfig.Tls_skip_verify,
		},
	)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultConfig.Address)
	client.SetToken("")
	return client, nil
}

func NewGoldfishVaultClient() (client *api.Client, err error) {
	if client, err = NewVaultClient(); err == nil {
		client.SetToken(vaultToken)
	}
	return client, err
}

func Bootstrap(wrappingToken string) error {
	if wrappingToken == "" {
		return errors.New("Wrapping token must be provided")
	}

	client, err := NewVaultClient()
	if err != nil {
		return err
	}

	// make a raw unwrap call. This will use the token as a header
	client.SetToken(wrappingToken)
	resp, err := client.Logical().Unwrap("")
	if err != nil {
		return errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}
	if resp == nil {
		return errors.New("Vault response was nil. Please revoke token.\n" +
			"If your vault cert is self-signed, you'll need to enable tls_skip_verify in goldfish config.")
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
	resp, err = client.Logical().Write(vaultConfig.Approle_login,
		map[string]interface{}{
			"role_id":   vaultConfig.Approle_id,
			"secret_id": secretID,
		})
	if err != nil {
		return err
	}

	// verify that the client token is sufficiently privileged
	vaultToken = resp.Auth.ClientToken
	if err := CheckServerRights(); err != nil {
		vaultToken = ""
		return err
	}

	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
	return nil
}

// similar to bootstrap function, but uses a raw token instead of an approle secret_id
// highly dangerous and not recommended unless approle is inaccessible
func BootstrapRaw() error {
	if vaultConfig.Raw_token == "" {
		return errors.New("Raw_token is not set in configuration")
	}

	// ensure the token has necessary rights
	vaultToken = vaultConfig.Raw_token
	if err := CheckServerRights(); err != nil {
		vaultToken = ""
		return err
	}

	// log this token's accessor
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}
	resp, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("Could not bootstrap, vault returned nil response")
	}

	log.Println("[INFO ]: Server token accessor:", resp.Data["accessor"])
	return nil
}

// check to ensure server's token has basic rights, and is able to read config path
func CheckServerRights() error {
	if vaultToken == "" {
		return errors.New("vault token must be set to check rights")
	}

	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}

	// verify server token can lookup self (this should be in default policy)
	if _, err := client.Auth().Token().LookupSelf(); err != nil {
		return err
	}
	// verify that the client token is renewable
	if err := renewServerToken(); err != nil {
		return err
	}
	// load runtime config once to make sure policy allows for it
	if err := loadConfigFromVault(vaultConfig.Runtime_config); err != nil {
		return err
	}

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
