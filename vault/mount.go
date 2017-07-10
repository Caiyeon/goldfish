package vault

import (
	"errors"

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
	if path == "" {
		return nil, errors.New("Empty mount name")
	}

	client, err := auth.Client()
	if err != nil {
		return nil, err
	}

	return client.Sys().MountConfig(path+"/")
}

func (auth AuthInfo) TuneMount(path string, config api.MountConfigInput) error {
	if path == "" {
		return errors.New("Empty mount name")
	}

	client, err := auth.Client()
	if err != nil {
		return err
	}

	return client.Sys().TuneMount(path+"/", config)
}
