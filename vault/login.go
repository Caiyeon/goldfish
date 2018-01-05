package vault

import (
	"errors"
	"strings"

	"github.com/hashicorp/vault/api"
)

// constructs a client with server's vault address and client access token
func (auth AuthInfo) Client() (client *api.Client, err error) {
	if client, err = NewVaultClient(); err == nil {
		client.SetToken(auth.ID)
	}
	return client, err
}

// verifies whether auth ID and password are valid
// if valid, creates a client access token and returns the metadata
func (auth *AuthInfo) Login() (map[string]interface{}, error) {
	client, err := NewVaultClient()
	if err != nil {
		return nil, err
	}
	client.SetToken("")

	// supported means there's a mapping to how the login should be performed
	t := strings.ToLower(auth.Type)
	key, exists := LoginMap[t]
	if !exists {
		return nil, errors.New("Unsupported authentication type: " + t)
	}

	// token logins don't require any writes to vault
	if t == "token" {
		client.SetToken(auth.ID)
	}

	// github logins are special: they only have one auth piece
	// these fields need to be swapped for the frontend to handle them easily
	if t == "github" {
		auth.Pass = auth.ID
		auth.ID = ""
	}

	// if logging in for the first time with these auth backends
	if t == "userpass" || t == "ldap" || t == "github" || t == "okta" {
		// fetch a client token by writing to vault auth backend
		loginPath := "auth/" + t + "/login/" + auth.ID

		// if auth has a different backend name, use that
		if auth.Path != "" {
			loginPath = "auth/" + auth.Path + "/login/" + auth.ID
		}

		resp, err := client.Logical().Write(
			loginPath,
			map[string]interface{}{
				key: auth.Pass,
			})
		if err != nil {
			return nil, err
		}
		// sanity check to make sure client token exists
		if resp.Auth == nil || resp.Auth.ClientToken == "" {
			return nil, errors.New("Unable to parse vault response")
		}
		// set the returned client token as the client's auth
		client.SetToken(resp.Auth.ClientToken)
	}

	// user must be able to lookup-self. This is in the default policy
	lookupResp, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return nil, err
	}

	// set auth type to token, so future requests don't need a login again
	auth.Type = "token"
	auth.ID = client.Token()
	auth.Pass = ""

	return lookupResp.Data, nil
}

func (auth AuthInfo) RenewSelf() (*api.Secret, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return client.Auth().Token().RenewSelf(0)
}

func (auth AuthInfo) LookupSelf() (*api.Secret, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return client.Auth().Token().LookupSelf()
}

// Logging in with different methods requires different secondary keys
var LoginMap = map[string]string{
    "token": "",
	"userpass": "password",
	"github": "token",
	"ldap": "password",
	"okta": "password",
}
