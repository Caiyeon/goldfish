package vault

import (
	"errors"
	"strings"
)

type LDAPUser struct {
	Policies []string
	Groups   []string
}

func (auth AuthInfo) ListLDAPGroups() (map[string][]string, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	resp, err := logical.List("auth/ldap/groups")
	if err != nil {
		return nil, err
	}
	groups, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to fetch LDAP group names")
	}

	results := make(map[string][]string)
	for _, g := range groups {
		group, ok := g.(string)
		if !ok {
			continue
		}
		results[group] = []string{}

		resp, err := logical.Read("auth/ldap/groups/" + group)
		if err == nil && resp != nil {
			if policies, ok := resp.Data["policies"]; ok {
				results[group] = strings.Split(policies.(string), ",")
			}
		}
	}

	return results, nil
}

func (auth AuthInfo) ListLDAPUsers() (map[string]*LDAPUser, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	resp, err := logical.List("auth/ldap/users")
	if err != nil {
		return nil, err
	}
	users, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to fetch LDAP usernames")
	}

	results := make(map[string]*LDAPUser)
	for _, u := range users {
		user, ok := u.(string)
		if !ok {
			continue
		}
		results[user] = &LDAPUser{}

		resp, err := logical.Read("auth/ldap/users/" + user)
		if err != nil || resp == nil {
			continue
		}
		if raw, ok := resp.Data["policies"]; ok {
			if policies, ok := raw.(string); ok {
				results[user].Policies = strings.Split(policies, ",")
			}
		}
		if raw, ok := resp.Data["groups"]; ok {
			if groups, ok := raw.(string); ok {
				results[user].Groups = strings.Split(groups, ",")
			}
		}
	}

	return results, nil
}
