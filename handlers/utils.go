package handlers

import (
	"github.com/hashicorp/vault/api"
)

// helper function to setup vault
func setupClient() error {
	client, err := api.NewClient(api.DefaultConfig())
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	if _, err = client.Auth().Token().LookupSelf(); err != nil {
		return err
	}
	vaultClient = client
	return nil
}
