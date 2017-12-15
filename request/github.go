package request

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"reflect"

	"github.com/caiyeon/goldfish/github"
	"github.com/caiyeon/goldfish/vault"
	"github.com/hashicorp/go-multierror"
	"github.com/fatih/structs"
)

type GithubRequest struct {
	Type          string
	CommitHash    string
	Changes       map[string]PolicyDiff
	Requester     string
	RequesterHash string
	Required      int
	Progress      int `hash:"ignore"`
}

type PolicyDiff struct {
	Previous string
	Proposed string
}

func (r GithubRequest) IsRootOnly() bool {
	return true
}

// verifies user can read all policies in the changes
func CreateGithubRequest(auth *vault.AuthInfo, raw map[string]interface{}) (*GithubRequest, error) {
	r := &GithubRequest{
		Changes: make(map[string]PolicyDiff),
	}
	r.Type = "github"

	if temp, ok := raw["commithash"]; ok {
		r.CommitHash, _ = temp.(string)
	}
	if r.CommitHash == "" {
		return nil, errors.New("'commithash' is required")
	}

	// collect requester's information
	self, err := auth.LookupSelf()
	if err != nil {
		return nil, err
	}
	if self == nil {
		return nil, errors.New("Could not confirm requester identity")
	}
	r.Requester = self.Data["display_name"].(string)
	r.RequesterHash = fmt.Sprintf("%x", sha256.Sum256([]byte(r.Requester)))

	// collect vault sys info
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return nil, err
	}
	r.Required = status.Required
	r.Progress = 0

	// fetch changes from github
	conf := vault.GetConfig()
	newPolicies, err := github.GetHCLFilesFromPath(
		conf.GithubAccessToken,
		conf.GithubRepoOwner,
		conf.GithubRepo,
		"",
		conf.GithubPoliciesPath,
		"",
		r.CommitHash,
	)
	if err != nil {
		// split by colon to prevent information disclosure with github api requests
		errtext := strings.Split(err.Error(), ":")
		return nil, errors.New(strings.Trim(errtext[len(errtext)-1], " "))
	}

	// fetch existing policies from vault
	currentPolicies, err := auth.ListPolicies()
	if err != nil {
		return nil, errors.New("Could not list existing policies: " + err.Error())
	}

	// for each hcl file from github, add an entry
	for name, future := range newPolicies {
		// verify user has rights to see policy
		current, err := auth.GetPolicy(name)
		if err != nil {
			return nil, errors.New("Could not read existing policy " + name + ": " + err.Error())
		}

		// if there is a difference, add it to the request changes
		if current != future {
			r.Changes[name] = PolicyDiff{
				Previous: current,
				Proposed: future,
			}
		}
	}

	// for each policy in vault that wasn't found on github, mark it as to be deleted
	for _, name := range currentPolicies {
		// a missing root or default policy is fine, don't delete either of these
		if name == "root" || name == "default" {
			continue
		}
		if _, ok := newPolicies[name]; !ok {
			// if policy exists in vault but not in github
			current, err := auth.GetPolicy(name)
			if err != nil {
				return nil, errors.New("Could not read existing policy " + name + ": " + err.Error())
			}
			r.Changes[name] = PolicyDiff{
				Previous: current,
				Proposed: "",
			}
		}
	}

	// if vault and github are identical, don't create the request in cubbyhole
	if len(r.Changes) == 0 {
		return nil, errors.New("No changes detected")
	}

	return r, nil
}

// verifies user can read all policies that will be changed
// if vault's policies changed in the meanwhile, progress will be reset
func (r *GithubRequest) Verify(auth *vault.AuthInfo) error {
	prevProgress := r.Progress

	// fetch a current copy of the diff between github and vault
	reqNow, err := CreateGithubRequest(auth, map[string]interface{}{
		"commithash": r.CommitHash,
	})
	if err != nil {
		return err
	}

	// compare stored vs new diffs
	if !reflect.DeepEqual(r.Changes, reqNow.Changes) {
		// if vault policies no longer match creation, reset progress and update change list
		r.Changes = reqNow.Changes
		r.Progress = 0
	}

	// check if vault key info is the same
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return err
	}
	if r.Required != status.Required {
		r.Progress = 0
		r.Required = status.Required
	}

	// if progress has been reset, purge unseal keys from cubbyhole
	if prevProgress != r.Progress {
		if _, err := vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + r.CommitHash); err != nil {
			return err
		}
	}

	return nil
}

// provides and unseal as an approval to a request
// if there are sufficient unseal tokens, attempt to roll the change
func (r *GithubRequest) Approve(hash string, unsealKey string) error {
	if unsealKey == "" {
		return errors.New("Unseal key cannot be empty")
	}

	// github requests don't rely on external provided hash
	hash = r.CommitHash

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
	r.Progress = r.Required
	defer vault.DeleteFromCubbyhole("requests/" + hash)
	defer rootAuth.RevokeSelf()

	// for each policy in diff, update it to the proposed copy
	var multierr error
	for name, diff := range r.Changes {
		if err := rootAuth.PutPolicy(name, diff.Proposed); err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	return multierr
}

// purges the request entry and unseal tokens from goldfish's cubbyhole
func (r *GithubRequest) Reject(auth *vault.AuthInfo, hash string) error {
	if _, err := vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash); err != nil {
		return err
	}
	if _, err := vault.DeleteFromCubbyhole("requests/" + hash); err != nil {
		return err
	}
	return nil
}
