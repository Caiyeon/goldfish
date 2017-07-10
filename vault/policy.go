package vault

import (
	"errors"
)

func (auth AuthInfo) ListPolicies() ([]string, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	return client.Sys().ListPolicies()
}

func (auth AuthInfo) GetPolicy(name string) (string, error) {
	client, err := auth.Client()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("Empty policy name")
	}
	return client.Sys().GetPolicy(name)
}

func (auth AuthInfo) DeletePolicy(name string) error {
	client, err := auth.Client()
	if err != nil {
		return err
	}
	if name == "" {
		return errors.New("Empty policy name")
	}
	return client.Sys().DeletePolicy(name)
}

func (auth AuthInfo) PutPolicy(name, rules string) error {
	client, err := auth.Client()
	if err != nil {
		return err
	}
	if name == "" {
		return errors.New("Empty policy name")
	}
	return client.Sys().PutPolicy(name, rules)
}
