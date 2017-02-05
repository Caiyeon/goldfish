package handlers

import (
	"errors"

	"github.com/hashicorp/vault/api"
	uuid "github.com/hashicorp/go-uuid"
)

// type AuthInfo struct {
// 	Type  string `json:"Type" form:"Type" query:"Type"`
// 	ID    string `json:"ID" form:"ID" query:"ID"`
// }

// returns err if authentication is invalid
// returns nil otherwise
func (auth AuthInfo) check() error {
	switch auth.Type{
	case "token":
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			return err
		}
		client.SetAddress(vaultAddress)
		client.SetToken(auth.ID)
		_, err = client.Auth().Token().LookupSelf()
		return err
	default:
		return errors.New("Unsupported authentication type")
	}
}

// stores auth details in a newly generated UUID path inside cubbyhole
// returns path (UUID) on success, error on failure
func (auth AuthInfo) store() (string, error) {
	path, err := uuid.GenerateUUID()
	if err != nil {
		return "", err
	}
	_, err = vaultClient.Logical().Write(
		"cubbyhole/sessions/" + path,
		map[string]interface{}{
			"Type": auth.Type,
			"ID":   auth.ID,
		})
	return path, err
}
