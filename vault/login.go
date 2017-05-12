package vault

import (
	"errors"
	"github.com/hashicorp/vault/api"
)

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

		// let future requests re-use the client token
		auth.Type = "token"
		auth.ID = resp.Auth.ClientToken
		auth.Pass = ""
		return lookupResp.Data, nil

	case "github":
		client.SetToken("")
		// fetch client access token by performing a login
		resp, err := client.Logical().Write("auth/github/login",
			map[string]interface{}{
				"token": auth.ID,
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

		// let future requests re-use the client token
		auth.Type = "token"
		auth.ID = resp.Auth.ClientToken
		return lookupResp.Data, nil

	case "ldap":
		client.SetToken("")
		resp, err := client.Logical().Write("auth/ldap/login/" + auth.ID,
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

		// let future requests re-use the client token
		auth.Type = "token"
		auth.ID   = resp.Auth.ClientToken
		auth.Pass = ""
		return lookupResp.Data, nil

	default:
		return nil, errors.New("Unsupported authentication type")
	}
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
