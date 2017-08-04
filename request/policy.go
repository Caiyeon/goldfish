package request

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	"github.com/caiyeon/goldfish/vault"
	"github.com/fatih/structs"
	"github.com/hashicorp/hcl"
	"github.com/mitchellh/hashstructure"
)

type PolicyRequest struct {
	Type          string
	PolicyName    string
	Previous      string
	Proposed      string
	Requester     string
	RequesterHash string
	Required      int
	Progress      int `hash:"ignore"`
}

func (r PolicyRequest) IsRootOnly() bool {
	return true
}

// verifies user can read policy, and that it hasn't changed since proposal
func (r *PolicyRequest) Verify(auth vault.AuthInfo) error {
	// verify new policy confirms to HCL formatting
	if _, err := hcl.Parse(r.Proposed); err != nil {
		return errors.New("Policy details cannot be parsed as HCL")
	}

	// verify user can read policy and it hasn't been changed
	policyCurrent, err := auth.GetPolicy(r.PolicyName)
	if err != nil {
		return err
	}
	if policyCurrent != r.Previous {
		return errors.New("Policy has been changed since request was made")
	}

	// verify it's a real change
	if r.Previous == r.Proposed {
		return errors.New("Policy details already match proposed change")
	}

	// if vault's key count has changed, the request is invalid
	if status, err := vault.GenerateRootStatus(); err != nil {
		return err
	} else if status.Required != r.Required {
		return errors.New("Request outdated due to vault rekey")
	}

	return nil
}

// constructs the request from limited fields and returns the hash
// raw must contain two keys: 'policyname' and 'rules'
func (r *PolicyRequest) Create(auth vault.AuthInfo, raw map[string]interface{}) (string, error) {
	// assert required fields
	r.Type = "policy"
	r.PolicyName = ""
	if temp, ok := raw["policyname"]; ok {
		r.PolicyName, _ = temp.(string)
	}
	if r.PolicyName == "" {
		return "", errors.New("'policyname' is required")
	}

	r.Proposed = ""
	if temp, ok := raw["rules"]; ok {
		r.Proposed, _ = temp.(string)
	}
	if r.Proposed == "" {
		return "", errors.New("No changes proposed")
	}
	if _, err := hcl.Parse(r.Proposed); err != nil {
		return "", errors.New("Policy must be HCL formatted")
	}

	// collect requester's information
	self, err := auth.LookupSelf()
	if err != nil {
		return "", err
	}
	if self == nil {
		return "", errors.New("Could not confirm requester identity")
	}
	r.Requester = self.Data["display_name"].(string)
	r.RequesterHash = fmt.Sprintf("%x", sha256.Sum256([]byte(r.Requester)))

	// verify user has access to read policy
	r.Previous, err = auth.GetPolicy(r.PolicyName)
	if err != nil {
		return "", err
	}
	if r.Previous == r.Proposed {
		return "", errors.New("Request contains no changes to policy")
	}

	// collect vault sys info
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return "", err
	}
	r.Required = status.Required
	r.Progress = 0

	// calculate and return hash
	hash_uint64, err := hashstructure.Hash(r, nil)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(hash_uint64, 16), nil
}

// provides an unseal token as an approval to a request
// if there are sufficient unseal tokens, attempt to roll the change
func (r *PolicyRequest) Approve(hash string, unsealKey string) error {
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

	// prepare cleanup
	defer vault.DeleteFromCubbyhole("requests/" + hash)
	defer rootAuth.RevokeSelf()

	// make requested change
	return rootAuth.PutPolicy(r.PolicyName, r.Proposed)
}

// purges the request entry and unseal tokens from goldfish's cubbyhole
func (r *PolicyRequest) Reject(auth vault.AuthInfo, hash string) error {
	if _, err := vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash); err != nil {
		return err
	}
	if _, err := vault.DeleteFromCubbyhole("requests/" + hash); err != nil {
		return err
	}
	return nil
}
