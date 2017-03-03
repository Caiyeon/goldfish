package vault

import (
	"errors"
)

func (auth AuthInfo) ListSecret(path string) (interface{}, error) {
	client, err := auth.Client()
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

func (auth AuthInfo) ReadSecret(path string) (interface{}, error) {
	client, err := auth.Client()
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