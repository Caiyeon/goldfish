package vault

import (
	"encoding/base64"
	"errors"

	"github.com/hashicorp/vault/api"
)

// zeros out credentials, call by defer
func (auth *AuthInfo) Clear() {
	auth.Type = ""
	auth.ID = ""
	auth.Pass = ""
}

func (auth AuthInfo) RevokeSelf() error {
	client, err := auth.Client()
	if err != nil {
		return err
	}
	return client.Auth().Token().RevokeSelf("")
}

// encrypt auth details with transit backend
func (auth *AuthInfo) EncryptAuth() error {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}

	c := GetConfig()

	resp, err := client.Logical().Write(
		c.TransitBackend + "/encrypt/" + c.ServerTransitKey,
		map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(auth.ID)),
		})
	if err != nil {
		return err
	}

	cipher, ok := resp.Data["ciphertext"].(string)
	if !ok {
		return errors.New("Failed type assertion of response to string")
	}

	auth.ID = cipher
	return nil
}

// decrypt auth details with transit backend
func (auth *AuthInfo) DecryptAuth() error {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}

	c := GetConfig()

	resp, err := client.Logical().Write(
		c.TransitBackend + "/decrypt/" + c.ServerTransitKey,
		map[string]interface{}{
			"ciphertext": auth.ID,
		})
	if err != nil {
		return err
	}

	b64, ok := resp.Data["plaintext"].(string)
	if !ok {
		return errors.New("Failed type assertion of response to string")
	}

	rawbytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	auth.ID = string(rawbytes)
	return nil
}

// returns a list of capabilities the current auth has on a given path
func (auth *AuthInfo) CapabilitiesSelf(path string) ([]string, error) {
	client, err := auth.Client()
	if err != nil {
		return []string{}, err
	}
	return client.Sys().CapabilitiesSelf(path)
}

func (auth AuthInfo) DeleteRaw(path string) (*api.Secret, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return client.Logical().Delete(path)
}
