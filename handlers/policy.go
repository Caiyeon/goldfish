package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/caiyeon/goldfish/github"
	"github.com/caiyeon/goldfish/slack"
	"github.com/caiyeon/goldfish/vault"
	"github.com/fatih/structs"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/vault/helper/xor"
	"github.com/labstack/echo"
	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
)

var reqsMutex = &sync.Mutex{}

type PolicyRequest struct {
	Policy        string
	Current       string
	New           string
	Requester     string
	RequesterHash string
	Required      int
	Progress      int `hash:"ignore"`
}

type PolicyDiff struct {
	Policy  string
	Current string
	New     string
}

func GetPolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// if policy is empty string, all policies will be fetched
		var result interface{}
		var err error
		policy := c.QueryParam("policy")
		if policy == "" {
			result, err = auth.ListPolicies()
		} else {
			result, err = auth.GetPolicy(policy)
		}

		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeletePolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch results
		if err := auth.DeletePolicy(c.QueryParam("policy")); err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "Policy deleted",
		})
	}
}

// Adds a policy request to cubbyhole, that can be rejected/approved later
// Requires requester to have read access to the policy's rule
func AddPolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		policy := c.QueryParam("policy")

		// check if user has access to policy
		policyOld, err := auth.GetPolicy(policy)
		if err != nil {
			return parseError(c, err)
		}

		// verify new policy conforms to HCL formatting
		policyNew := c.FormValue("rules")
		if _, err := hcl.Parse(policyNew); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Policy must be HCL formatted",
			})
		}

		if policyOld == policyNew {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Policy request is identical to current",
			})
		}

		// collect non-dangerous identifying data on requester
		self, err := auth.LookupSelf()
		if err != nil {
			return parseError(c, err)
		}

		// get number of unseal keys required to generate root token
		status, err := vault.GenerateRootStatus()
		if err != nil {
			return parseError(c, err)
		}

		// construct request
		requester, ok := self.Data["display_name"].(string)
		if !ok {
			return c.JSON(http.StatusInternalServerError, H{
				"error": "Could not parse requester display name",
			})
		}
		accessor, ok := self.Data["accessor"].(string)
		if !ok {
			return c.JSON(http.StatusInternalServerError, H{
				"error": "Could not hash requester token accessor",
			})
		}
		request := PolicyRequest{
			Policy:        policy,
			Current:       policyOld,
			New:           policyNew,
			Requester:     requester,
			RequesterHash: fmt.Sprintf("%x", sha256.Sum256([]byte(accessor))),
			Required:      status.Required,
			Progress:      0,
		}

		// hash request structure
		hash_uint64, err := hashstructure.Hash(request, nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, H{
				"error": "Could not hash request",
			})
		}
		hash := strconv.FormatUint(hash_uint64, 16)

		// write to cubbyhole with details (lock is required)
		reqsMutex.Lock()
		_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(request))
		if err != nil {
			return parseError(c, err)
		}
		reqsMutex.Unlock()

		// if config has a slack webhook, send the hash (aka change ID) to the channel
		conf := vault.GetConfig()
		if conf.SlackWebhook != "" {
			// send a message using webhook
			err = slack.PostMessageWebhook(
				conf.SlackChannel,
				"A new policy change request has been submitted",
				"Change ID: \n*"+hash+"*",
				conf.SlackWebhook,
			)
			// change request is fine, just let the frontend know it wasn't slack'd
			if err != nil {
				return c.JSON(http.StatusOK, H{
					"result": hash,
					"error":  "Could not send to slack webhook",
				})
			}
		}

		// return hash
		return c.JSON(http.StatusOK, H{
			"result": hash,
			"error":  "",
		})
	}
}

// Searches a policy request from cubbyhole
// Requires requester to have read access to the policy's rule
func GetPolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		switch c.QueryParam("type") {
		case "changeid":
			if c.QueryParam("id") == "" {
				return c.JSON(http.StatusBadRequest, H{
					"error": "id param is required",
				})
			}
			return getPolicyRequestByChangeID(c, auth, c.QueryParam("id"))

		case "commit":
			if c.QueryParam("sha") == "" {
				return c.JSON(http.StatusBadRequest, H{
					"error": "sha param is required",
				})
			}
			return getPolicyRequestByCommitHash(c, auth, c.QueryParam("sha"))

		default:
			return c.JSON(http.StatusBadRequest, H{
				"error": "Type must be either changeid or commit",
			})
		}
	}
}

func getPolicyRequestByChangeID(c echo.Context, auth *vault.AuthInfo, hash string) error {
	// fetch change from cubbyhole
	resp, err := vault.ReadFromCubbyhole("requests/" + hash)
	if err != nil {
		return parseError(c, err)
	}
	if resp == nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Change ID not found",
		})
	}

	// decode map to struct
	var request PolicyRequest
	err = mapstructure.Decode(resp.Data, &request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Request appears to be malformed",
		})
	}

	// verify current user has rights to see policy
	policyCurrent, err := auth.GetPolicy(request.Policy)
	if err != nil {
		return parseError(c, err)
	}

	// verify hash
	statusCode, err := verifyRequest(request, hash, policyCurrent)
	if err != nil {
		return c.JSON(statusCode, H{
			"error": err.Error(),
		})
	}

	// if vault has been re-keyed, the request is invalid
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return parseError(c, err)
	}
	if request.Required != status.Required {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Request outdated due to vault rekey",
		})
	}

	return c.JSON(http.StatusOK, H{
		"result":   request,
		"progress": request.Progress,
		"required": request.Required,
	})
}

func getPolicyRequestByCommitHash(c echo.Context, auth *vault.AuthInfo, hash string) error {
	changes, err := compareGithubVault(auth, hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": err.Error(),
		})
	}

	// check progress and total unseals required
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return parseError(c, err)
	}
	cubbyhole, err := vault.ReadFromCubbyhole("unseal_wrapping_tokens/" + hash)
	if err != nil {
		return parseError(c, err)
	}

	progress := 0
	if cubbyhole != nil && cubbyhole.Data != nil {
		if wrapping_tokens, ok := cubbyhole.Data["wrapping_tokens"]; ok {
			progress = len(strings.Split(wrapping_tokens.(string), ";"))
		}
	}

	return c.JSON(http.StatusOK, H{
		"result":   changes,
		"progress": progress,
		"required": status.Required,
	})
}

func compareGithubVault(auth *vault.AuthInfo, hash string) ([]PolicyDiff, error) {
	// fetch change from github
	conf := vault.GetConfig()
	newPolicies, err := github.GetHCLFilesFromPath(
		conf.GithubAccessToken,
		conf.GithubRepoOwner,
		conf.GithubRepo,
		conf.GithubTargetBranch,
		conf.GithubPoliciesPath,
		vault.GithubCurrentCommit,
		hash,
	)
	if err != nil {
		// split by colon to prevent information disclosure with github api requests
		errtext := strings.Split(err.Error(), ":")
		return nil, errors.New(strings.Trim(errtext[len(errtext)-1], " "))
	}

	currentPolicies, err := auth.ListPolicies()
	if err != nil {
		return nil, errors.New("Could not list existing policies")
	}

	changes := make([]PolicyDiff, 0)

	// for each hcl file in github folder, add an entry
	for name, future := range newPolicies {
		diff := PolicyDiff{Policy: name}

		// verify current user has rights to see policy
		current, err := auth.GetPolicy(name)
		if err != nil {
			return nil, errors.New("Could not read existing policy")
		}

		// github package already verified that future string is hcl-formatted
		if current != future {
			diff.Current = current
			diff.New = future
			changes = append(changes, diff)
		}
	}

	// for each current policy that doesn't exist in github, mark it as would be deleted
	for _, name := range currentPolicies {
		if name == "root" || name == "default" {
			continue
		}
		if _, ok := newPolicies[name]; !ok {
			// policy exists in vault but not in github
			diff := PolicyDiff{Policy: name}
			current, err := auth.GetPolicy(name)
			if err != nil {
				return nil, errors.New("Could not read existing policy")
			}
			diff.Current = current
			diff.New = ""
			changes = append(changes, diff)
		}
	}

	return changes, nil
}

// Provides an unseal token for a policy request
// If enough tokens are reached, a root token generation and policy change is attempted
func UpdatePolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		unsealKey := c.FormValue("unseal")
		if unsealKey == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Unseal param required",
			})
		}

		switch c.QueryParam("type") {
		case "changeid":
			if c.QueryParam("id") == "" {
				return c.JSON(http.StatusBadRequest, H{
					"error": "id param required",
				})
			}
			return updatePolicyRequestByChangeID(c, auth, c.QueryParam("id"), unsealKey)

		case "commit":
			if c.QueryParam("sha") == "" {
				return c.JSON(http.StatusBadRequest, H{
					"error": "sha param required",
				})
			}
			return updatePolicyRequestByCommitHash(c, auth, c.QueryParam("sha"), unsealKey)

		default:
			return c.JSON(http.StatusBadRequest, H{
				"error": "type must be either changeid or commit",
			})
		}
	}
}

func updatePolicyRequestByChangeID(c echo.Context, auth *vault.AuthInfo, hash string, unsealKey string) error {
	// full lock is required to prevent a race condition in providing unseal keys
	reqsMutex.Lock()
	defer reqsMutex.Unlock()

	// fetch change from cubbyhole
	resp, err := vault.ReadFromCubbyhole("requests/" + hash)
	if err != nil {
		return parseError(c, err)
	}
	if resp == nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Change ID not found",
		})
	}

	// decode map to struct
	var request PolicyRequest
	err = mapstructure.Decode(resp.Data, &request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Change appears to be malformed",
		})
	}

	// verify current user has rights to see policy
	policyCurrent, err := auth.GetPolicy(request.Policy)
	if err != nil {
		return parseError(c, err)
	}

	// verify hash
	statusCode, err := verifyRequest(request, hash, policyCurrent)
	if err != nil {
		return c.JSON(statusCode, H{
			"error": err.Error(),
		})
	}

	// if vault has been re-keyed, the request is invalid
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return parseError(c, err)
	}
	if request.Required != status.Required {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Request outdated by vault rekey",
		})
	}

	// count how many unseals are entered so far
	wrappingTokens := []string{}
	if request.Progress > 0 {
		resp, err := vault.ReadFromCubbyhole("unseal_wrapping_tokens/" + hash)
		if err != nil {
			return parseError(c, err)
		}
		wrappingTokens = strings.Split(resp.Data["wrapping_tokens"].(string), ";")
	}

	// wrap the unseal token
	newWrappingToken, err := vault.WrapData("60m", map[string]interface{}{
		"unseal_token": unsealKey,
	})
	if err != nil {
		return parseError(c, err)
	}

	// add the new wrapping token to the slice
	wrappingTokens = append(wrappingTokens, newWrappingToken)

	// if there aren't enough unseals yet
	if len(wrappingTokens) < request.Required {
		// store the wrapping tokens back in cubbyhole
		_, err = vault.WriteToCubbyhole("unseal_wrapping_tokens/"+hash,
			map[string]interface{}{
				"wrapping_tokens": strings.Trim(strings.Join(strings.Fields(fmt.Sprint(wrappingTokens)), ";"), "[]"),
			})
		if err != nil {
			return parseError(c, err)
		}

		// store progress in request too
		request.Progress = len(wrappingTokens)
		_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(request))
		if err != nil {
			return parseError(c, err)
		}

		// return progress
		return c.JSON(http.StatusOK, H{
			"progress": len(wrappingTokens),
			"required": status.Required,
		})
	}

	// if we got here, it means there are enough unseals to attempt root generation
	// so if we exit after this point, progress must be reset
	request.Progress = 0
	defer vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash)

	// unwrap all the unseal tokens
	unseals := []string{}
	for _, wrappingToken := range wrappingTokens {
		data, err := vault.UnwrapData(wrappingToken)
		if err != nil {
			return parseError(c, err)
		}
		if unseal, ok := data["unseal_token"]; !ok {
			return c.JSON(http.StatusInternalServerError, H{
				"error": "One of the wrapping tokens timed out. Progress reset",
			})
		} else {
			unseals = append(unseals, unseal.(string))
		}
	}

	// start a root generation
	randomBytes, err := uuid.GenerateRandomBytes(16)
	if err != nil {
		return parseError(c, err)
	}
	otp := base64.StdEncoding.EncodeToString(randomBytes)
	status, err = vault.GenerateRootInit(otp)
	if err != nil {
		return parseError(c, err)
	}

	// feed unseal tokens
	if status.EncodedRootToken == "" {
		for _, s := range unseals {
			status, err = vault.GenerateRootUpdate(s, status.Nonce)
			// an error likely means one of the unseals was not valid
			if err != nil {
				// delete root generation process
				if err := vault.GenerateRootCancel(); err != nil {
					return parseError(c, err)
				}
				// inform user that request unseals have been reset
				return c.JSON(http.StatusBadRequest, H{
					"error": "At least one unseal key was invalid. Progress has been reset.",
				})
			}
		}
	}

	// sanity check
	if status.EncodedRootToken == "" {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Root generation failed. Was vault re-keyed just now?",
		})
	}

	// decode root token
	tokenBytes, err := xor.XORBase64(status.EncodedRootToken, otp)
	if err != nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Could not decode root token. Please search and revoke it manually",
		})
	}
	token, err := uuid.FormatUUID(tokenBytes)
	if err != nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": "Could not decode root token. Please search and revoke it manually",
		})
	}

	// perform policy change with generated root token
	var rootauth = &vault.AuthInfo{
		Type: "token",
		ID:   token,
	}

	// ensure generated root token is revoked, and cubbyhole data is purged
	defer vault.DeleteFromCubbyhole("requests/" + hash)
	defer rootauth.RevokeSelf()

	// make requested change
	err = rootauth.PutPolicy(request.Policy, request.New)
	if err != nil {
		return parseError(c, err)
	}

	// confirm changes have been applied
	policyNow, err := auth.GetPolicy(request.Policy)
	if err != nil {
		return parseError(c, err)
	}

	return c.JSON(http.StatusOK, H{
		"result": policyNow,
	})
}

func updatePolicyRequestByCommitHash(c echo.Context, auth *vault.AuthInfo, hash string, unsealKey string) error {
	// full lock is required to prevent a race condition in providing unseal keys
	reqsMutex.Lock()
	defer reqsMutex.Unlock()

	// fetch difference in policies
	changes, err := compareGithubVault(auth, hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, H{
			"error": err.Error(),
		})
	}

	// check progress
	cubbyhole, err := vault.ReadFromCubbyhole("unseal_wrapping_tokens/" + hash)
	if err != nil {
		return parseError(c, err)
	}

	// wrap this unseal token
	newWrappingToken, err := vault.WrapData("60m", map[string]interface{}{
		"unseal_token": unsealKey,
	})
	if err != nil {
		return parseError(c, err)
	}

	// see the current progress on this request
	wrappingTokens := []string{}
	if cubbyhole != nil && cubbyhole.Data != nil {
		if temp, ok := cubbyhole.Data["wrapping_tokens"]; ok {
			wrappingTokens = strings.Split(temp.(string), ";")
		}
	}

	// add the newly wrapped token to the slice
	wrappingTokens = append(wrappingTokens, newWrappingToken)

	// if there aren't enough unseals yet, store them all and return progress and required
	status, err := vault.GenerateRootStatus()
	if err != nil {
		return parseError(c, err)
	}
	if len(wrappingTokens) < status.Required {
		_, err = vault.WriteToCubbyhole("unseal_wrapping_tokens/"+hash,
			map[string]interface{}{
				"wrapping_tokens": strings.Trim(strings.Join(strings.Fields(fmt.Sprint(wrappingTokens)), ";"), "[]"),
			})
		if err != nil {
			return parseError(c, err)
		}
		return c.JSON(http.StatusOK, H{
			"progress": len(wrappingTokens),
			"required": status.Required,
		})
	}

	// if we got here, it means there are enough unseals to attempt root generation
	// so if we exit after this point, progress must be reset
	defer vault.DeleteFromCubbyhole("unseal_wrapping_tokens/" + hash)

	// unwrap all the unseal tokens
	unseals := []string{}
	for _, wrappingToken := range wrappingTokens {
		data, err := vault.UnwrapData(wrappingToken)
		if err != nil {
			return parseError(c, err)
		}
		if unseal, ok := data["unseal_token"]; !ok {
			return c.JSON(http.StatusInternalServerError, H{
				"error": "One of the wrapping tokens timed out. Progress reset",
			})
		} else {
			unseals = append(unseals, unseal.(string))
		}
	}

	// start a root generation
	randomBytes, err := uuid.GenerateRandomBytes(16)
	if err != nil {
		return parseError(c, err)
	}
	otp := base64.StdEncoding.EncodeToString(randomBytes)
	status, err = vault.GenerateRootInit(otp)
	if err != nil {
		return parseError(c, err)
	}

	// feed unseal tokens
	if status.EncodedRootToken == "" {
		for _, s := range unseals {
			status, err = vault.GenerateRootUpdate(s, status.Nonce)
			// an error likely means one of the unseals was not valid
			if err != nil {
				// delete root generation process
				if err := vault.GenerateRootCancel(); err != nil {
					return parseError(c, err)
				}
				// inform user that request unseals have been reset
				return c.JSON(http.StatusBadRequest, H{
					"error": "At least one unseal key was invalid. Progress has been reset.",
				})
			}
		}
	}

	// sanity check
	if status.EncodedRootToken == "" {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Root generation failed. Was vault re-keyed just now?",
		})
	}

	// decode root token
	tokenBytes, err := xor.XORBase64(status.EncodedRootToken, otp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Could not decode root token. Please search and revoke manually",
		})
	}
	token, err := uuid.FormatUUID(tokenBytes)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, H{
			"error": "Could not decode root token. Please search and revoke manually",
		})
	}

	// perform policy change with generated root token
	var rootauth = &vault.AuthInfo{
		Type: "token",
		ID:   token,
	}

	// ensure generated root token is revoked
	defer rootauth.RevokeSelf()

	// make all requested changes
	for _, policyDiff := range changes {
		err = rootauth.PutPolicy(policyDiff.Policy, policyDiff.New)
		if err != nil {
			return parseError(c, err)
		}
	}

	return c.JSON(http.StatusOK, H{
		"result": "Complete",
	})
}

// Anyone that is able to read the policy is able to delete change requests for that policy
func DeletePolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// full lock is required to prevent a race condition in providing unseal keys
		reqsMutex.Lock()
		defer reqsMutex.Unlock()

		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch change from cubbyhole
		hash := c.Param("id")
		resp, err := vault.ReadFromCubbyhole("requests/" + hash)
		if err != nil {
			return parseError(c, err)
		}
		if resp == nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Change ID not found",
			})
		}

		// fetch policy name from change
		policyName, ok := resp.Data["Policy"]
		if !ok {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Change appears to be malformed",
			})
		}

		// verify current user has rights to see policy
		_, err = auth.GetPolicy(policyName.(string))
		if err != nil {
			return parseError(c, err)
		}

		// purge change related data from cubbyhole
		_, err = vault.DeleteFromCubbyhole("unseals/" + hash)
		if err != nil {
			return parseError(c, err)
		}
		_, err = vault.DeleteFromCubbyhole("requests/" + hash)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "Request deleted",
		})
	}
}

func verifyRequest(request PolicyRequest, hash string, policyCurrent string) (int, error) {
	hash_uint64, err := hashstructure.Hash(request, nil)
	if err != nil || strconv.FormatUint(hash_uint64, 16) != hash {
		return http.StatusBadRequest, errors.New("Hashes do not match")
	}

	// verify that policy has not been changed since change was requested
	if policyCurrent != request.Current {
		return http.StatusBadRequest, errors.New("Policy has been changed since request was made")
	}

	// verify new policy conforms to HCL formatting
	if _, err := hcl.Parse(request.New); err != nil {
		return http.StatusBadRequest, errors.New("Policy details cannot be parsed as HCL")
	}

	// verify change is still... well, a change.
	if policyCurrent == request.New {
		return http.StatusBadRequest, errors.New("Policy details already match proposed change")
	}

	return http.StatusOK, nil
}
