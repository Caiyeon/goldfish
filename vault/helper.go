package vault

import (
	"errors"

	"github.com/hashicorp/vault/api"
)

func VaultHealth() (*api.HealthResponse, error) {
	client, err := NewVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Sys().Health()
}

// lookup current root generation status
func GenerateRootStatus() (*api.GenerateRootStatusResponse, error) {
	client, err := NewVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Sys().GenerateRootStatus()
}

func GenerateRootInit(otp string) (*api.GenerateRootStatusResponse, error) {
	client, err := NewVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Sys().GenerateRootInit(otp, "")
}

func GenerateRootUpdate(shard, nonce string) (*api.GenerateRootStatusResponse, error) {
	client, err := NewVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Sys().GenerateRootUpdate(shard, nonce)
}

func GenerateRootCancel() error {
	client, err := NewVaultClient()
	if err != nil {
		return err
	}
	return client.Sys().GenerateRootCancel()
}

func WriteToCubbyhole(name string, data map[string]interface{}) (interface{}, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Logical().Write("cubbyhole/"+name, data)
}

func ReadFromCubbyhole(name string) (*api.Secret, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Logical().Read("cubbyhole/" + name)
}

func DeleteFromCubbyhole(name string) (*api.Secret, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return nil, err
	}
	return client.Logical().Delete("cubbyhole/" + name)
}

func WrapData(wrapttl string, data map[string]interface{}) (string, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return "", err
	}

	client.SetWrappingLookupFunc(func(operation, path string) string {
		return wrapttl
	})

	resp, err := client.Logical().Write("/sys/wrapping/wrap", data)
	if err != nil {
		return "", err
	}
	return resp.WrapInfo.Token, nil
}

func UnwrapData(wrappingToken string) (map[string]interface{}, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return nil, err
	}

	// make a raw unwrap call. This will use the token as a header
	resp, err := client.Logical().Unwrap(wrappingToken)
	if err != nil {
		return nil, errors.New("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}
	return resp.Data, nil
}

func LookupSelf() (map[string]interface{}, error) {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().Read("/auth/token/lookup-self")
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
