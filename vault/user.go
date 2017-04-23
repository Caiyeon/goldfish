package vault

import (
	"encoding/json"
	"errors"
)

func (auth AuthInfo) ListUsers(backend string) (interface{}, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	switch backend {
	case "token":
		// get a list of token accessors
		resp, err := logical.List("auth/token/accessors")
		if err != nil {
			return nil, err
		}
		accessors, ok := resp.Data["keys"].([]interface{})
		if !ok {
			return nil, errors.New("Failed to convert response")
		}

		// fetch each token's details
		tokens := make([]interface{}, len(accessors))
		for i, accessor := range accessors {
			resp, err := logical.Write("auth/token/lookup-accessor",
				map[string]interface{}{
					"accessor": accessor,
				})
			// error may occur if accessor expired. Simply ignore it.
			if err == nil {
				tokens[i] = resp.Data
			}
		}
		return tokens, nil

	case "userpass":
		type User struct {
			Name     string
			TTL      int
			Max_TTL  int
			Policies string
		}

		// get a list of usernames
		resp, err := logical.List("auth/userpass/users")
		if err != nil {
			return nil, err
		}
		usernames, ok := resp.Data["keys"].([]interface{})
		if !ok {
			return nil, errors.New("Failed to convert response")
		}

		// fetch each user's details
		users := make([]User, len(usernames))
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

	case "approle":
		type Role struct {
			Roleid             string
			Token_TTL          int
			Token_max_TTL      int
			Secret_id_TTL      int
			Secret_id_num_uses int
			Policies           []string
			Period             int
			Bind_secret_id     bool
			Bound_cidr_list    string
		}

		// get a list of roles
		resp, err := logical.List("auth/approle/role")
		if err != nil {
			return nil, err
		}
		rolenames, ok := resp.Data["keys"].([]interface{})
		if !ok {
			return nil, errors.New("Failed to convert response")
		}

		// fetch each role's details
		roles := make([]Role, len(rolenames))
		for i, role := range rolenames {
			roles[i].Roleid = role.(string)
			resp, err := logical.Read("auth/approle/role/" + roles[i].Roleid)
			if err == nil {
				if b, err := json.Marshal(resp.Data); err == nil {
					json.Unmarshal(b, &roles[i])
				}
			}
		}
		return roles, nil

	default:
		return nil, errors.New("Unsupported user listing type")
	}
}

func (auth AuthInfo) DeleteUser(backend string, deleteID string) error {
	client, err := auth.Client()
	if err != nil {
		return err
	}
	logical := client.Logical()

	if deleteID == "" {
		return errors.New("Invalid deletion ID")
	}

	switch backend {
	case "token":
		_, err := logical.Write("/auth/token/revoke-accessor/" + deleteID, nil)
		return err

	case "userpass":
		_, err := logical.Delete("/auth/userpass/users/" + deleteID)
		return err

	case "approle":
		_, err := logical.Delete("/auth/approle/role/" + deleteID)
		return err

	default:
		return errors.New("Unsupported user deletion type")
	}
}
