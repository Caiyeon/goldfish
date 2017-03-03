package vault

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"flag"
	"net/http"
	"io/ioutil"

	"github.com/hashicorp/vault/api"
)

// for authenticating this web server with vault
var vaultAddress = ""
var vaultToken = ""
var vaultClient *api.Client

// transit key for server use only
// encrypts and decrypts user's authentication info in cookie
var serverTransitKey = "goldfish"

// transit key for user usage
// used for encrypting and decrypting strings in tools/transit
var userTransitKey = "usertransit"

type AuthInfo struct {
	Type string `json:"Type" form:"Type" query:"Type"`
	ID   string `json:"ID" form:"ID" query:"ID"`
}

func init() {
	gob.Register(&AuthInfo{})

	// to do: change token to approle
	flag.StringVar(&vaultAddress, "addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&vaultToken, "token", "", "Vault token")
	flag.Parse()
	if vaultAddress == "" || vaultToken == "" {
		panic("Invalid vault credentials")
	}

	// set up web server's vault client
	client, err := api.NewClient(api.DefaultConfig())
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	if _, err = client.Auth().Token().LookupSelf(); err != nil {
		panic(err)
	}
	vaultClient = client
}

// zeros out credentials, call by defer
func (auth *AuthInfo) Clear() {
	auth.Type = ""
	auth.ID = ""
}

// returns a constructed client with the server's vaultaddress and provided ID
func (auth AuthInfo) Client() (*api.Client, error) {
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
func (auth *AuthInfo) EncryptAuth() error {
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
func (auth *AuthInfo) DecryptAuth() error {
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