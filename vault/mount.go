package vault

import (
	"github.com/hashicorp/vault/api"
)

// returns list of current mounts, if authorized
func (auth AuthInfo) ListMounts() (map[string]*api.MountOutput, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	return client.Sys().ListMounts()
}

func (auth AuthInfo) GetMount(path string) (*api.MountConfigOutput, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	return client.Sys().MountConfig(path + "/")
}

func (auth AuthInfo) TuneMount(path string, config api.MountConfigInput) error {
	client, err := auth.Client()
	if err != nil {
		return err
	}

	return client.Sys().TuneMount(path+"/", config)
}
