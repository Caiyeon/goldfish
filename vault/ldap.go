package vault

import (
	"errors"
	"strings"
)

type LDAPGroup struct {
	Name     string
	Policies []string
}

type LDAPUser struct {
	Name     string
	Policies []string
	Groups   []string
}

func (auth AuthInfo) ListLDAPGroups() ([]LDAPGroup, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	resp, err := logical.List("auth/ldap/groups")
	if err != nil {
		return nil, err
	}
	raw, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to fetch LDAP group names")
	}

	// ignore any group names that somehow can't be type asserted to string
	var groups []string
	for _, each := range raw {
		if group, ok := each.(string); ok {
			groups = append(groups, group)
		}
	}

	results := make([]LDAPGroup, len(groups))
	for i, group := range groups {
		results[i] = LDAPGroup{
			Name: group,
		}
		// fetch group's policies
		resp, err := logical.Read("auth/ldap/groups/" + group)
		if err == nil && resp != nil {
			if policies, ok := resp.Data["policies"]; ok {
				if p, ok := policies.(string); ok {
					results[i].Policies = strings.Split(p, ",")
				}
			}
		}
	}

	return results, nil
}

func (auth AuthInfo) ListLDAPUsers() ([]LDAPUser, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	resp, err := logical.List("auth/ldap/users")
	if err != nil {
		return nil, err
	}
	raw, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to fetch LDAP usernames")
	}

	// ignore any user names that somehow can't be type asserted to string
	var users []string
	for _, each := range raw {
		if user, ok := each.(string); ok {
			users = append(users, user)
		}
	}

	results := make([]LDAPUser, len(users))
	for i, user := range users {
		results[i] = LDAPUser{
			Name: user,
		}
		// fetch user's policies and groups
		resp, err := logical.Read("auth/ldap/users/" + user)
		if err != nil || resp == nil {
			continue
		}
		if raw, ok := resp.Data["policies"]; ok {
			if policies, ok := raw.(string); ok {
				results[i].Policies = strings.Split(policies, ",")
			}
		}
		if raw, ok := resp.Data["groups"]; ok {
			if groups, ok := raw.(string); ok {
				results[i].Groups = strings.Split(groups, ",")
			}
		}
	}

	return results, nil
}
