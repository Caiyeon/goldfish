package vault

import (
	"errors"

	api "github.com/hashicorp/vault/vault"
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

func (auth AuthInfo) PolicyCapabilities(policyName, path string) ([]string, error) {
	client, err := auth.Client()
	if err != nil {
		return []string{}, err
	}

	// the "root" policy is unchangeable and unremovable, so short circuit
	if policyName == "root" {
		return []string{api.RootCapability}, nil
	}

	// fetch policy rules
	rules, err := client.Sys().GetPolicy(policyName)
	if err != nil {
		return []string{}, err
	}

	// construct ACL
	policy, err := api.ParseACLPolicy(rules)
	if err != nil {
		return []string{}, err
	}
	acl, err := api.NewACL([]*api.Policy{policy})
	if err != nil {
		return []string{}, err
	}

	// read capabilities of policy and return
	return acl.Capabilities(path), nil
}
