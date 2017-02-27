package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/hashicorp/vault/api"
)

// vault transit key, used for encrypting and decrypting user's auth cookie
// to invalidate all auth cookies, simply rotate this key in vault
var serverTransitKey = "goldfish"

// vault transit key, used for encrypting and decrypting strings in tools/transit
// separate from the server's transit key so that each can be rotated independently
var userTransitKey = "usertransit"

// zeros out credentials. Should be called via defer auth.clear()
func (auth *AuthInfo) clear() {
	auth.Type = ""
	auth.ID = ""
}

// returns a constructed client with the server's vaultaddress and provided ID
func (auth AuthInfo) client() (*api.Client, error) {
	switch auth.Type {
	case "token":
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return nil, err
		}
		client.SetAddress(vaultAddress)
		client.SetToken(auth.ID)
		_, err = client.Auth().Token().LookupSelf()
		return client, err
	// only tokens are supported for now
	default:
		return nil, errors.New("Unsupported authentication type")
	}
}

// encrypt auth details with transit backend
func (auth *AuthInfo) encrypt() error {
	resp, err := vaultClient.Logical().Write(
		"transit/encrypt/" + serverTransitKey,
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
func (auth *AuthInfo) decrypt() error {
	resp, err := vaultClient.Logical().Write(
		"transit/decrypt/" + serverTransitKey,
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

func (auth AuthInfo) listusers(backend string) (interface{}, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	switch backend {
	case "token":
		// get a list of token accessors
		resp, err := logical.List("auth/token/accessors")
		if err != nil {
			return nil, err
		}
		accessors, ok := resp.Data["keys"].([]interface{})
		if !ok {
			return nil, errors.New("Failed to convert response")
		}

		// fetch each token's details
		tokens := make([]interface{}, len(accessors))
		for i, accessor := range accessors {
			resp, err := logical.Write("auth/token/lookup-accessor",
				map[string]interface{}{
					"accessor": accessor,
				})
			// error may occur if accessor expired. Simply ignore it.
			if err == nil {
				tokens[i] = resp.Data
			}
		}
		return tokens, nil

	case "userpass":
		type User struct {
			Name     string
			TTL      int
			Max_TTL  int
			Policies string
		}

		// get a list of usernames
		resp, err := logical.List("auth/userpass/users")
		if err != nil {
			return nil, err
		}
		usernames, ok := resp.Data["keys"].([]interface{})
		if !ok {
			return nil, errors.New("Failed to convert response")
		}

		// fetch each user's details
		users := make([]User, len(usernames))
		for i, username := range usernames {
			users[i].Name = username.(string)
			resp, err := logical.Read("auth/userpass/users/" + users[i].Name)
			if err == nil {
				if b, err := json.Marshal(resp.Data); err == nil {
					json.Unmarshal(b, &users[i])
				}
			}
		}
		return users, nil

	default:
		return nil, errors.New("Unsupported user listing type")
	}
}

func (auth AuthInfo) deleteuser(backend string, deleteID string) error {
	client, err := auth.client()
	if err != nil {
		return err
	}
	logical := client.Logical()

	if deleteID == "" {
		return errors.New("Invalid deletion ID")
	}

	switch backend {
	case "token":
		_, err := logical.Write("/auth/token/revoke-accessor/" + deleteID, nil)
		return err

	case "userpass":
		_, err := logical.Delete("/auth/userpass/users/" + deleteID)
		return err

	default:
		return errors.New("Unsupported user deletion type")
	}
}

func (auth AuthInfo) listpolicies() ([]string, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}
	return client.Sys().ListPolicies()
}

func (auth AuthInfo) getpolicy(name string) (string, error) {
	client, err := auth.client()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("Invalid policy name")
	}
	return client.Sys().GetPolicy(name)
}

func (auth AuthInfo) deletepolicy(name string) error {
	client, err := auth.client()
	if err != nil {
		return err
	}
	if name == "" {
		return errors.New("Invalid policy name")
	}
	return client.Sys().DeletePolicy(name)
}

// encrypt given string with user transit key
func (auth AuthInfo) encryptstring(plaintext string) (string, error) {
	client, err := auth.client()
	if err != nil {
		return "", err
	}

	resp, err := client.Logical().Write(
		"transit/encrypt/" + userTransitKey,
		map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte(plaintext)),
		})
	if err != nil {
		return "", err
	}

	cipher, ok := resp.Data["ciphertext"].(string)
	if !ok {
		return "", errors.New("Failed type assertion of response to string")
	}

	return cipher, nil
}

// decrypt auth details with transit backend
func (auth AuthInfo) decryptstring(cipher string) (string, error) {
	client, err := auth.client()
	if err != nil {
		return "", err
	}

	resp, err := client.Logical().Write(
		"transit/decrypt/" + userTransitKey,
		map[string]interface{}{
			"ciphertext": cipher,
		})
	if err != nil {
		return "", err
	}

	b64, ok := resp.Data["plaintext"].(string)
	if !ok {
		return "", errors.New("Failed type assertion of response to string")
	}

	rawbytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}

	return string(rawbytes), nil
}

// returns list of current mounts, if authorized
func (auth AuthInfo) listmounts() (map[string]*api.MountOutput, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}

	return client.Sys().ListMounts()
}

func (auth AuthInfo) getmount(path string) (*api.MountConfigOutput, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}

	return client.Sys().MountConfig(path + "/")
}

func (auth AuthInfo) configmount(path string, config api.MountConfigInput) error {
	client, err := auth.client()
	if err != nil {
		return err
	}

	return client.Sys().TuneMount(path + "/", config)
}

func (auth AuthInfo) listpath(path string) (interface{}, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().List(path)
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Data == nil {
		// invalid handler (i.e. invalid request)
		return nil, errors.New("Invalid path")
	} else {
		return resp.Data["keys"], nil
	}
}

func (auth AuthInfo) readpath(path string) (interface{}, error) {
	client, err := auth.client()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().Read(path)
	if err != nil {
		return nil, err
	}

	if resp == nil {
		// invalid handler (i.e. invalid request)
		return nil, errors.New("Invalid path")
	} else {
		return resp.Data, nil
	}
}