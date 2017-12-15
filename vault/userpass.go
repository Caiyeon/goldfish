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

	// fetch each user's details
	users := make([]UserpassUser, len(usernames))
	for i, username := range usernames {
		users[i].Name = username.(string)
		resp, err := logical.Read("auth/userpass/users/" + users[i].Name)
		if err == nil {
			if b, err := json.Marshal(resp.Data); err == nil {
				json.Unmarshal(b, &users[i])
			}
		}
	}
	return users, nil
}
