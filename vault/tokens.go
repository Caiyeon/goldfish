package vault

import (
	"errors"
	"strings"

	"github.com/hashicorp/vault/api"
)

func (auth AuthInfo) GetTokenAccessors() ([]interface{}, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().List("auth/token/accessors")
	if err != nil {
		return nil, err
	}

	accessors, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to fetch token accessors")
	}

	return accessors, nil
}

func (auth AuthInfo) LookupTokenByAccessor(accs string) ([]interface{}, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	// accessors should be comma delimited
	accessors := strings.Split(accs, ",")
	if len(accessors) == 1 && accessors[0] == "" {
		return nil, errors.New("No accessors provided")
	}

	// excessive numbers of tokens are not allowed, to avoid stress on vault
	if len(accessors) > 500 {
		return nil, errors.New("Maximum number of accessors: 500")
	}

	// for each accessor, lookup details
	tokens := make([]interface{}, len(accessors))
	for i, _ := range tokens {
		resp, err := logical.Write("auth/token/lookup-accessor",
			map[string]interface{}{
				"accessor": accessors[i],
			})
		// error may occur if accessor was invalid or expired, simply ignore it
		if err == nil {
			tokens[i] = resp.Data
		}
	}
	return tokens, nil
}

func (auth AuthInfo) RevokeTokenByAccessor(acc string) error {
	client, err := auth.Client()
	if err != nil {
		return err
	}
	logical := client.Logical()

	_, err = logical.Write("/auth/token/revoke-accessor/"+acc, nil)
	return err
}

func (auth AuthInfo) CreateToken(opts *api.TokenCreateRequest, orphan bool,
	rolename string, wrapttl string) (*api.Secret, error) {

	if orphan && rolename != "" {
		return nil, errors.New("Orphan and role are mutually exclusive parameters")
	}

	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	// if requester wants response wrapped
	if wrapttl != "" {
		client.SetWrappingLookupFunc(func(operation, path string) string {
			return wrapttl
		})
	}

	if orphan {
		return client.Auth().Token().CreateOrphan(opts)
	} else if rolename != "" {
		return client.Auth().Token().CreateWithRole(opts, rolename)
	}
	return client.Auth().Token().Create(opts)
}

func (auth AuthInfo) ListRoles() (interface{}, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().List("/auth/token/roles")
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return resp.Data["keys"], nil
}
func (auth AuthInfo) GetRole(rolename string) (interface{}, error) {
	if rolename == "" {
		return nil, errors.New("Empty rolename")
	}

	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	resp, err := client.Logical().Read("/auth/token/roles/" + rolename)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
