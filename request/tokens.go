package request

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/caiyeon/goldfish/vault"
	"github.com/fatih/structs"
    "github.com/hashicorp/vault/api"
	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
)

type TokenRequest struct {
	Type           string
    Orphan         string
    Wrap_ttl       string
    Role           string
    CreateRequest  *api.TokenCreateRequest
	CreateResponse *api.Secret
	Requester      string
	RequesterHash  string
	Required       int
	Progress       int `hash:"ignore"`
}

func (r TokenRequest) IsRootOnly() bool {
	return false
}

// constructs the request from limited fields and returns the hash
// raw must contain key: 'wrap_ttl', and can contain 'orphan', 'role'
func CreateTokenRequest(auth *vault.AuthInfo, raw map[string]interface{}) (*TokenRequest, string, error) {
	r := &TokenRequest{}
	r.Type = "token"

    // decode raw map to token creation request
    if temp, ok := raw["create_request"]; ok {
        m, ok := temp.(map[string]interface{})
		if !ok {
            return nil, "", errors.New("'create_request' is a wrong format")
        }

        decoder, err := mapstructure.NewDecoder(
            &mapstructure.DecoderConfig{
        		Metadata: nil,
        		Result:   &r.CreateRequest,
        		TagName:  "json",
        	},
        )
        if err != nil {
    		return nil, "", errors.New("Could not decode mapstructure: " + err.Error())
    	}

        err = decoder.Decode(m)
        if err != nil {
            return nil, "", errors.New("Could not decode mapstructure: " + err.Error())
        }

        // created token would be revoked when the generated root token is revoked,
        // so no_parent needs to be set
        r.CreateRequest.NoParent = true
    } else {
        return nil, "", errors.New("'create_request' is required")
    }

    if temp, ok := raw["wrap_ttl"]; ok {
		r.Wrap_ttl, _ = temp.(string)
	}
	if r.Wrap_ttl == "" {
		return nil, "", errors.New("'wrap_ttl' is required")
	}
    if temp, err := strconv.Atoi(r.Wrap_ttl); err != nil {
        return nil, "", errors.New("'wrap_ttl' could not be translated to integer")
    } else if temp < 1 {
        return nil, "", errors.New("'wrap_ttl' must be greater than 1")
    }

    if temp, ok := raw["orphan"]; ok {
        if r.Orphan, ok = temp.(string); !ok {
            return nil, "", errors.New("'orphan' must be in string format")
        }
    }
    if r.Orphan != "" && r.Orphan != "true" && r.Orphan != "false" {
        return nil, "", errors.New("'orphan' must be empty, 'true', or 'false'")
    }

    if temp, ok := raw["role"]; ok {
        if r.Role, ok = temp.(string); !ok {
            return nil, "", errors.New("'role' must be in string format")
        }
    }

	if r.Orphan == "true" && r.Role != "" {
		return nil, "", errors.New("'role' and 'orphan' fields are mutually exclusive")
	}

	// collect requester's information
	self, err := auth.LookupSelf()
	if err != nil {
		return nil, "", err
	}
	if self == nil {
		return nil, "", errors.New("Could not confirm requester identity")
	}
	r.Requester = self.Data["display_name"].(string)
	r.RequesterHash = fmt.Sprintf("%x", sha256.Sum256([]byte(r.Requester)))

	// verify user has at least default policy
    if _, err := auth.Login(); err != nil {
        return nil, "", err
    }

	// collect vault sys info
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return nil, "", err
	}
	r.Required = status.Required
	r.Progress = 0

	// calculate hash
	hash_uint64, err := hashstructure.Hash(r, nil)
	if err != nil {
		return nil, "", err
	}
	hash := strconv.FormatUint(hash_uint64, 16)
	if hash == "" {
		return nil, "", errors.New("Failed to hash request")
	}

	return r, hash, nil
}

// verifies user can read the role if request contains one, and at least lookup self
func (r *TokenRequest) Verify(auth *vault.AuthInfo) error {
    // verify user has at least default policy
    if _, err := auth.Login(); err != nil {
        return err
    }

    // if request has a specific role, the approver should be able to read it
    if r.Role != "" {
        if _, err := auth.GetRole(r.Role); err != nil {
            return err
        }
    }

	// if vault's key count has changed, the request is invalid
	if status, err := vault.GenerateRootStatus(); err != nil {
		return err
	} else if status.Required != r.Required {
		return errors.New("Request outdated due to vault rekey")
	}

	return nil
}

// provides an unseal token as an approval to a request
// if there are sufficient unseal tokens, attempt to roll the change
func (r *TokenRequest) Approve(hash string, unsealKey string) error {
	if unsealKey == "" {
		return errors.New("Unseal key cannot be empty")
	}

	// append unseal key to cubbyhole
	wrappingTokens, err := appendUnseal(hash, unsealKey)
	if err != nil {
		return err
	}

	// if there aren't enough unseals yet, update progress
	if r.Required > len(wrappingTokens) {
		r.Progress = len(wrappingTokens)
		_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(r))
		return err
	}

	// prepare cleanup
	r.Progress = 0
	defer vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash)

	// unwrap the unseal tokens
	unseals, err := unwrapUnseals(wrappingTokens)
	if err != nil {
		vault.WriteToCubbyhole("requests/"+hash, structs.Map(r))
		return err
	}

	// generate root token
	rootToken, err := generateRootToken(unseals)
	if err != nil {
		vault.WriteToCubbyhole("requests/"+hash, structs.Map(r))
		return err
	}
	var rootAuth = &vault.AuthInfo{
		Type: "token",
		ID:   rootToken,
	}

	// update progress
	r.Progress = r.Required

	// prepare cleanup
	defer vault.DeleteFromCubbyhole("requests/" + hash)
	defer rootAuth.RevokeSelf()

    // make requested change
    if r.CreateResponse, err = rootAuth.CreateToken(
		r.CreateRequest,
		r.Orphan == "true",
		r.Role,
		r.Wrap_ttl,
	); err != nil {
		return err
	}

	return nil
}

// purges the request entry and unseal tokens from goldfish's cubbyhole
func (r *TokenRequest) Reject(auth *vault.AuthInfo, hash string) error {
	if _, err := vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash); err != nil {
		return err
	}
	if _, err := vault.DeleteFromCubbyhole("requests/" + hash); err != nil {
		return err
	}
	return nil
}
