package vault

import (
	"encoding/json"
	"errors"
)

type UserpassUser struct {
	Name     string
	TTL      int
	Max_TTL  int
	Policies string
}

func (auth AuthInfo) ListUserpassUsers() ([]UserpassUser, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	// get a list of usernames
	resp, err := logical.List("auth/userpass/users")
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Data == nil {
		return []UserpassUser{}, nil
	}

	usernames, ok := resp.Data["keys"].([]interface{})
	if !ok {
		return nil, errors.New("Failed to convert response")
	}
	
	results := make([]UserpassUser, len(usernames))
	for i, username := range usernames {
		results[i].Name = username.(string)

		// fetch user's policies and groups
		resp, err := logical.Read("auth/userpass/users/" + results[i].Name)
		if err != nil || resp == nil {
			continue
		}

		if raw, ok := resp.Data["policies"]; ok {
			// vault v0.8.3 and higher returns an array of strings
			if policies, ok := raw.([]interface{}); ok {
				for _, p := range policies {
					if s, ok := p.(string); ok {
						results[i].Policies = append(results[i].Policies, s)
					                            }
				                            }
        	}
            }
	}
	
	return results, nil
	
 }
	

