package vault

import (
	"errors"
	"log"
	"time"
	"sync"

	"github.com/caiyeon/goldfish/config"
	"github.com/hashicorp/vault/api"
)

type AuthInfo struct {
	Type string `json:"type" form:"Type" query:"Type"`
	ID   string `json:"ID" form:"ID" query:"ID"`
	Pass string `json:"password" form:"Password" query:"Password"`
	Path string `json:"path" form:"Path" query:"Path"`
}

var (
	vaultConfig          config.VaultConfig
	vaultToken           string
	vaultTokenLock       = new(sync.RWMutex)

	errorChannel         = make(chan error, 1)
	stopChannelConfig    = make(chan struct{}, 1)
	stopChannelRenew     = make(chan struct{}, 1)
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
	vaultTokenLock.RLock()
	defer vaultTokenLock.RUnlock()
	return vaultToken != ""
}

func unbootstrap() {
	if !Bootstrapped() {
		return
	}

	// clear token
	vaultTokenLock.Lock()
	defer vaultTokenLock.Unlock()
	vaultToken = ""

	// stop background processes
	go func() {
		stopChannelConfig <- struct{}{}
		stopChannelRenew <- struct{}{}
	}()
}

func SetConfig(c *config.VaultConfig) {
	vaultConfig = *c
}

func NewVaultClient() (*api.Client, error) {
	config := api.DefaultConfig()
	err := config.ConfigureTLS(
		&api.TLSConfig{
			Insecure: vaultConfig.Tls_skip_verify,
			CACert: vaultConfig.CA_cert,
			CAPath: vaultConfig.CA_path,
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
	if !Bootstrapped() {
		return nil, errors.New("Goldfish is not bootstrapped yet!")
	}

	if client, err = NewVaultClient(); err == nil {
		vaultTokenLock.RLock()
		defer vaultTokenLock.RUnlock()
		client.SetToken(vaultToken)
	}
	return client, err
}

func Bootstrap(wrappingToken string) error {
	if Bootstrapped() {
		return errors.New("Already bootstrapped. Re-bootstrapping is not supported at the moment.")
	}
	if wrappingToken == "" {
		return errors.New("Wrapping token must be provided")
	}

	// unwrap token
	client, err := NewVaultClient()
	if err != nil {
		return err
	}
	client.SetToken(wrappingToken)

	resp, err := client.Logical().Unwrap("")
	if err != nil {
		return errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}
	if resp == nil {
		return errors.New("Vault response was nil.\n" +
			"If your vault cert is self-signed, you need to enable tls_skip_verify in goldfish config.")
	}

	var potentialToken string

	// if this is an approle secret_id:
	if resp.Data != nil {
		// parse secret_id
		raw, ok := resp.Data["secret_id"]
		if !ok {
			return errors.New("Could not find secret_id in wrapped token. Was it wrapped properly?")
		}
		secretID, ok := raw.(string)
		if !ok {
			return errors.New("Could not assert secret_id as string. Was it wrapped properly?")
		}

		// login with approle
		resp2, err := client.Logical().Write(vaultConfig.Approle_login,
			map[string]interface{}{
				"role_id":   vaultConfig.Approle_id,
				"secret_id": secretID,
			})
		if err != nil {
			return err
		}
		if resp2 == nil || resp2.Auth == nil {
			return errors.New("Error fetching client token in approle login. Is the role configured properly?")
		}
		potentialToken = resp2.Auth.ClientToken

	// if this is a regular wrapped token (not approle)
	} else if resp.Auth != nil {
		potentialToken = resp.Auth.ClientToken

	} else {
		return errors.New("Wrapped secret is neither approle nor token. Aborting.")
	}

	// verify token has sufficient rights
	acc, err := VerifyTokenRights(potentialToken)
	if err != nil {
		return err
	}

	// lock in token
	vaultTokenLock.Lock()
	vaultToken = potentialToken
	vaultTokenLock.Unlock()

	// load runtime config
	if err := loadConfigFromVault(vaultConfig.Runtime_config); err != nil {
		vaultTokenLock.Lock()
		vaultToken = ""
		vaultTokenLock.Unlock()
		return err
	}

	// notify user of the accessor so it can be revoked if needed
	log.Println("[INFO ]: Successfully bootstrapped. Server token accessor:", acc)

	// start background processes for renewing settings and server token
	go loadConfigEvery(time.Minute, vaultConfig.Runtime_config)
	go renewServerTokenEvery(time.Hour)

	return nil
}

// similar to bootstrap function, but uses a raw token instead of an approle secret_id
// highly dangerous and not recommended to be called externally unless approle is inaccessible
func BootstrapRaw(token string) error {
	if Bootstrapped() {
		return errors.New("Already bootstrapped. Re-bootstrapping is not supported at the moment.")
	}
	if token == "" {
		return errors.New("No token provided")
	}

	// ensure the token has necessary rights
	acc, err := VerifyTokenRights(token)
	if err != nil {
		return err
	}

	// lock in token
	vaultTokenLock.Lock()
	vaultToken = token
	vaultTokenLock.Unlock()

	// load runtime config
	if err := loadConfigFromVault(vaultConfig.Runtime_config); err != nil {
		vaultTokenLock.Lock()
		vaultToken = ""
		vaultTokenLock.Unlock()
		return err
	}

	// notify user of the accessor so it can be revoked if needed
	log.Println("[INFO ]: Server token accessor:", acc)

	// start background processes for renewing settings and server token
	go loadConfigEvery(time.Minute, vaultConfig.Runtime_config)
	go renewServerTokenEvery(time.Hour)

	return nil
}

// check to ensure server's token has basic rights, and is able to read config path
func VerifyTokenRights(token string) (accessor string, err error) {
	client, err := NewVaultClient()
	if err != nil {
		return "", err
	}
	client.SetToken(token)

	// verify token can lookup self (this should be in default policy)
	if resp, err := client.Auth().Token().LookupSelf(); err != nil {
		return "", errors.New("Could not lookup self. Goldfish requires default policy. Error: " + err.Error())
	} else if resp == nil {
		return "", errors.New("Could not lookup self. Goldfish requires default policy. Response from vault was nil")
	} else {
		// if lookup succeeded, record the token's accessor
		accessor = resp.Data["accessor"].(string)
	}

	// verify token is renewable
	if resp, err := client.Auth().Token().RenewSelf(0); err != nil {
		return "", errors.New("Could not renew token. Goldfish requires a renewable role/token. Error: " + err.Error())
	} else if resp == nil {
		return "", errors.New("Could not renew token. Goldfish requires a renewable role/token. Response from vault was nil")
	}

	// verify token can read runtime settings
	resp, err := client.Logical().Read(vaultConfig.Runtime_config)
	if err != nil {
		return "", errors.New("Could not read runtime settings. Error: " + err.Error())
	} else if resp == nil {
		return "", errors.New("Could not read runtime settings. Vault response was nil. Error: " + err.Error())
	}

	// good enough
	return accessor, nil
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
		select {
			default:
				errorChannel <- loadConfigFromVault(configPath)
			case <-stopChannelConfig:
				return
		}
	}
}

func renewServerTokenEvery(interval time.Duration) {
	for {
		time.Sleep(interval)
		select {
			default:
				errorChannel <- renewServerToken()
			case <-stopChannelRenew:
				return
		}
	}
}

func renewServerToken() error {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}
	resp, err := client.Auth().Token().RenewSelf(0)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("Could not renew token... response from vault was nil")
	}
	log.Println("[INFO ]: Server token renewed")
	return nil
}
