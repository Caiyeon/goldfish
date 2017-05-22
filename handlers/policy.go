package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caiyeon/goldfish/slack"
	"github.com/caiyeon/goldfish/vault"

	"github.com/fatih/structs"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/vault/helper/xor"

	"github.com/labstack/echo"

	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
)

type PolicyRequest struct {
	Policy        string
	Current       string
	New           string
	Requester     string
	RequesterHash string
	Required      int
	Progress      int    `hash:"ignore"`
}

func GetPolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

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
			return logError(c, err.Error(), "Internal error")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeletePolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		if err := auth.DeletePolicy(c.QueryParam("policy")); err != nil {
			return logError(c, err.Error(), "Internal error")
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
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		policy := c.QueryParam("policy")

		// check if user has access to policy
		policyOld, err := auth.GetPolicy(policy)
		if err != nil {
			return logError(c, err.Error(), "Could not read existing policy")
		}

		// verify new policy conforms to HCL formatting
		policyNew := c.FormValue("rules")
		if _, err := hcl.Parse(policyNew); err != nil {
			return logError(c, err.Error(), "Could not parse proposed policy rules")
		}

		if policyOld == policyNew {
			return logError(c, "", "No changes detected")
		}

		// collect non-dangerous identifying data on requester
		self, err := auth.LookupSelf()
		if err != nil {
			return logError(c, err.Error(), "Failed to perform lookupself on requester token")
		}

		// get number of unseal keys required to generate root token
		status, err := vault.GenerateRootStatus()
		if err != nil {
			return logError(c, err.Error(), "Could not check root generation status")
		}

		// construct request
		requester, ok := self.Data["display_name"].(string)
		if !ok {
			return logError(c, err.Error(), "Could not fetch requester display name")
		}
		accessor, ok := self.Data["accessor"].(string)
		if !ok {
			return logError(c, err.Error(), "Could not fetch requester accessor")
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
			return logError(c, err.Error(), "Could not hash request. Unsafe; aborting.")
		}
		hash := strconv.FormatUint(hash_uint64, 16)

		// write to cubbyhole with details
		_, err = vault.WriteToCubbyhole("requests/" + hash, structs.Map(request))
		if err != nil {
			return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
		}

		// if config has a slack webhook, send the hash (aka change ID) to the channel
		conf := vault.GetConfig()
		if webhook := conf.SlackWebhook; webhook != "" {
			// send a message using webhook
			err = slack.PostMessageWebhook(
				conf.SlackChannel,
				"A new policy change request has been submitted",
				"Change ID: \n*" + hash + "*",
				webhook,
			)
			// change request is fine, just let the frontend know it wasn't slack'd
			if err != nil {
				return c.JSON(http.StatusOK, H{
					"result": hash,
					"error": err.Error(),
				})
			}
		}

		// return hash
		return c.JSON(http.StatusOK, H{
			"result": hash,
			"error": "",
		})
	}
}

// Searches a policy request from cubbyhole
// Requires requester to have read access to the policy's rule
func GetPolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch change from cubbyhole
		hash := c.QueryParam("id")
		resp, err := vault.ReadFromCubbyhole("requests/" + hash)
		if err != nil {
			return logError(c, err.Error(), "Change ID not found")
		}
		if resp == nil {
			return logError(c, "", "Change ID not found")
		}

		// decode map to struct
		var request PolicyRequest
		err = mapstructure.Decode(resp.Data, &request)
		if err != nil {
			return logError(c, err.Error(), "Change appears to be malformed")
		}

		// verify current user has rights to see policy
		policyCurrent, err := auth.GetPolicy(request.Policy)
		if err != nil {
			return logError(c, err.Error(), "Could not read existing policy")
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
			return logError(c, err.Error(), "Could not check root generation status")
		}
		if request.Required != status.Required {
			return logError(c, "", "Number of unseal tokens required differs. Aborting")
		}

		// return request
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))
		return c.JSON(http.StatusOK, H{
			"result": request,
		})
	}
}

// Provides an unseal token for a policy request
// If enough tokens are reached, a root token generation and policy change is attempted
func UpdatePolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch change from cubbyhole
		hash := c.Param("id")
		resp, err := vault.ReadFromCubbyhole("requests/" + hash)
		if err != nil {
			return logError(c, err.Error(), "Change ID not found")
		}
		if resp == nil {
			return logError(c, "", "Change ID not found")
		}

		unsealKey := c.FormValue("unseal")
		if unsealKey == "" {
			return logError(c, err.Error(), "Must provide unseal key")
		}

		// decode map to struct
		var request PolicyRequest
		err = mapstructure.Decode(resp.Data, &request)
		if err != nil {
			return logError(c, err.Error(), "Change appears to be malformed")
		}

		// verify current user has rights to see policy
		policyCurrent, err := auth.GetPolicy(request.Policy)
		if err != nil {
			return logError(c, err.Error(), "Could not read existing policy")
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
			return logError(c, err.Error(), "Could not check root generation status")
		}
		if request.Required != status.Required {
			return logError(c, "", "Number of unseal tokens required differs. Aborting")
		}

		// count how many unseals are entered so far
		wrappingTokens := []string{}
		if request.Progress > 0 {
			resp, err := vault.ReadFromCubbyhole("unseal_wrapping_tokens/" + hash)
			if err != nil {
				return logError(c, err.Error(), "Could not read cubbyhole")
			}
			wrappingTokens = strings.Split(resp.Data["wrapping_tokens"].(string), ";")
		}

		// wrap the unseal token
		newWrappingToken, err := vault.WrapData("60m", map[string]interface{}{
			"unseal_token": unsealKey,
		})
		if err != nil {
			return logError(c, err.Error(), "Could not wrap unseal token")
		}

		// add the new wrapping token to the slice
		wrappingTokens = append(wrappingTokens, newWrappingToken)

		// if there aren't enough unseals yet
		if len(wrappingTokens) < request.Required {
			// store the wrapping tokens back in cubbyhole
			_, err = vault.WriteToCubbyhole("unseal_wrapping_tokens/" + hash,
				map[string]interface{}{
					"wrapping_tokens": strings.Trim(strings.Join(strings.Fields(fmt.Sprint(wrappingTokens)), ";"), "[]"),
				})
			if err != nil {
				return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
			}

			// store progress in request too
			request.Progress = len(wrappingTokens)
			_, err = vault.WriteToCubbyhole("requests/" + hash, structs.Map(request))
			if err != nil {
				return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
			}

			// return progress
			return c.JSON(http.StatusOK, H{
				"progress": len(wrappingTokens),
			})
		}

		// if we got here, it means there are enough unseals to attempt root generation
		// so if we exit after this point, progress must be reset
		request.Progress = 0
		defer vault.DeleteFromCubbyhole("unseals/" + hash)
		defer func(hash string, request PolicyRequest) {
			_, _ = vault.WriteToCubbyhole("requests/" + hash, structs.Map(request))
		}(hash, request)

		// unwrap all the unseal tokens
		unseals := []string{}
		for _, wrappingToken := range(wrappingTokens) {
			data, err := vault.UnwrapData(wrappingToken)
			if err != nil {
				return logError(c, err.Error(), "Wrapping token for unseal key seems to have timed out")
			}
			if unseal, ok := data["unseal_token"]; !ok {
				return logError(c, err.Error(), "Wrapping token for unseal key seems to have timed out")
			} else {
				unseals = append(unseals, unseal.(string))
			}
		}

		// start a root generation
		otp := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16))
		status, err = vault.GenerateRootInit(otp)
		if err != nil {
			return logError(c, err.Error(), "Another root generation is in progress. All unseal tokens have been erased.")
		}

		// feed unseal tokens
		if status.EncodedRootToken == "" {
			for _, s := range(unseals) {
				status, err = vault.GenerateRootUpdate(s, status.Nonce)
				// an error likely means one of the unseals was not valid
				if err != nil {
					// delete root generation process
					if err := vault.GenerateRootCancel(); err != nil {
						return logError(c, err.Error(), "At least one unseal key was invalid. Could not revert root generation!")
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
			return logError(c, err.Error(), "Could not decode root token. Please search and revoke it")
		}
		token, err := uuid.FormatUUID(tokenBytes)
		if err != nil {
			return logError(c, err.Error(), "Could not decode root token. Please search and revoke it")
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
			return logError(c, err.Error(), "Could not change policy")
		}

		// confirm changes have been applied
		policyNow, err := auth.GetPolicy(request.Policy)
		if err != nil {
			return logError(c, err.Error(), "Could not read policy after change")
		}

		// return request
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))
		return c.JSON(http.StatusOK, H{
			"result": policyNow,
		})
	}
}

// Anyone that is able to read the policy is able to delete change requests for that policy
func DeletePolicyRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch change from cubbyhole
		hash := c.Param("id")
		resp, err := vault.ReadFromCubbyhole("requests/" + hash)
		if err != nil {
			return logError(c, err.Error(), "Change ID not found")
		}
		if resp == nil {
			return logError(c, "", "Change ID not found")
		}

		// fetch policy name from change
		policyName, ok := resp.Data["Policy"]
		if !ok {
			return logError(c, err.Error(), "Change appears to be malformed")
		}

		// verify current user has rights to see policy
		_, err = auth.GetPolicy(policyName.(string))
		if err != nil {
			return logError(c, err.Error(), "Read permissions for policy are required to reject a change")
		}

		// purge change related data from cubbyhole
		_, err = vault.DeleteFromCubbyhole("unseals/" + hash)
		if err != nil {
			return logError(c, err.Error(), "Could not delete from cubbyhole")
		}
		_, err = vault.DeleteFromCubbyhole("requests/" + hash)
		if err != nil {
			return logError(c, err.Error(), "Could not delete from cubbyhole")
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
