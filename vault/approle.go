package vault

import (
	"encoding/json"
	"errors"
)

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

func (auth AuthInfo) ListApproleRoles() ([]Role, error) {
	client, err := auth.Client()
	if err != nil {
		return nil, err
	}
	logical := client.Logical()

	// get a list of roles
	resp, err := logical.List("auth/approle/role")
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Data == nil {
		return []Role{}, nil
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
}
