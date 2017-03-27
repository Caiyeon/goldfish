package vault

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"

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
	Type     string `json:"Type" form:"Type" query:"Type"`
	ID       string `json:"ID" form:"ID" query:"ID"`
	Pass     string `json:"password" form:"Password" query:"Password"`
}

func init() {
	// for gorilla securecookie to encode and decode
	gob.Register(&AuthInfo{})

	// read address and wrapping token inputs
	var wrappingToken, roleID, path string
	flag.StringVar(&vaultAddress, "addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&wrappingToken, "token", "", "Wrapping token that should contain a secret_id")
	flag.StringVar(&roleID, "role_id", "goldfish", "The role_id the secret_id was generated from")
	flag.StringVar(&path, "approle_path", "auth/approle/login", "The login path of the mount e.g. 'auth/approle/login'")
	flag.Parse()
	if vaultAddress == "" || wrappingToken == "" {
		panic("Provide credentials via -addr and -token (wrapping token only)")
	}

	// set up vault client
	client, err := api.NewClient(api.DefaultConfig())
	client.SetAddress(vaultAddress)
	client.SetToken(wrappingToken)

	// make a raw unwrap call. This will use the token as a header
	resp, err := client.Logical().Unwrap("")
	if err != nil {
		panic("Failed to unwrap provided token, revoke it if possible\nReason:" + err.Error())
	}

	// verify that a secret_id was wrapped
	secretID, ok := resp.Data["secret_id"].(string)
	if !ok {
		panic("Failed to unwrap provided token, revoke it if possible")
	}

	// fetch vault token with secret_id
	resp, err = client.Logical().Write(path,
		map[string]interface{}{
			"role_id":   roleID,
			"secret_id": secretID,
		})
	if err != nil {
		panic(err)
	}

	// verify that the secret_id is valid
	log.Println(resp.Auth.ClientToken)
	vaultToken = resp.Auth.ClientToken
	client.SetToken(vaultToken)
	if _, err = client.Auth().Token().LookupSelf(); err != nil {
		panic(err)
	}
	vaultClient = client

	// report back the accessor so it may be safekept
	log.Println("[INFO ]: Server token accessor:", resp.Auth.Accessor)
}

// zeros out credentials, call by defer
func (auth *AuthInfo) Clear() {
	auth.Type = ""
	auth.ID = ""
	auth.Pass = ""
}

// constructs a client with server's vault address and client access token
func (auth AuthInfo) Client() (*api.Client, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)
	client.SetToken(auth.ID)
	_, err = client.Auth().Token().LookupSelf()
	return client, err
}

// verifies whether auth ID and password are valid
// if valid, creates a client access token and returns the metadata
func (auth *AuthInfo) Login() (map[string]interface{}, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetAddress(vaultAddress)

	switch auth.Type {
	case "token":
		client.SetToken(auth.ID)
		resp, err := client.Auth().Token().LookupSelf()
		if err != nil {
			return nil, err
		}
		return resp.Data, nil

	case "userpass":
		client.SetToken("")
		// fetch client access token by performing a login
		resp, err := client.Logical().Write("auth/userpass/login/" + auth.ID,
			map[string]interface{}{
				"password": auth.Pass,
			})
		if err != nil {
			return nil, err
		}
		if resp.Auth == nil || resp.Auth.ClientToken == "" {
			return nil, errors.New("Unable to parse vault response")
		}

		client.SetToken(resp.Auth.ClientToken)
		lookupResp, err := client.Auth().Token().LookupSelf()
		if err != nil {
			return nil, err
		}

		// if the login was valid, set auth to the access token
		auth.Type = "token"
		auth.ID = resp.Auth.ClientToken
		auth.Pass = ""
		return lookupResp.Data, nil

	default:
		return nil, errors.New("Unsupported authentication type")
	}
}

// encrypt auth details with transit backend
func (auth *AuthInfo) EncryptAuth() error {
	resp, err := vaultClient.Logical().Write(
		"transit/encrypt/"+serverTransitKey,
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
		"transit/decrypt/"+serverTransitKey,
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
