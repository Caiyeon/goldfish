package vault

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/vault/api"
)

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

func VaultHealth() (string, error) {
	resp, err := http.Get(vaultAddress + "/v1/sys/health")
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// lookup current root generation status
func GenerateRootStatus() (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootStatus()
}

func GenerateRootInit(otp string) (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootInit(otp, "")
}

func GenerateRootUpdate(shard, nonce string) (*api.GenerateRootStatusResponse, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootUpdate(shard, nonce)
}

func GenerateRootCancel() error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	client.SetAddress(vaultAddress)
	return client.Sys().GenerateRootCancel()
}

func WriteToCubbyhole(name string, data map[string]interface{}) (interface{}, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Write("cubbyhole/" + name, data)
}

func ReadFromCubbyhole(name string) (*api.Secret, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Read("cubbyhole/" + name)
}

func DeleteFromCubbyhole(name string) (*api.Secret, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	return vaultClient.Logical().Delete("cubbyhole/" + name)
}
