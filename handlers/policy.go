package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caiyeon/goldfish/slack"
	"github.com/caiyeon/goldfish/vault"

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
	Policy    string
	Current   string
	New       string
	Requester string
	Required  int
	Progress  int    `hash:"ignore"`
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

		// construct request solely for hashing purposes
		request := PolicyRequest{
			Policy:    policy,
			Current:   policyOld,
			New:       policyNew,
			Requester: self.Data["display_name"].(string),
			Required:  status.Required,
			Progress:  0,
		}

		// hash structure
		hash_uint64, err := hashstructure.Hash(request, nil)
		if err != nil {
			return logError(c, err.Error(), "Could not hash request. Unsafe; aborting.")
		}
		hash := strconv.FormatUint(hash_uint64, 16)

		// write to cubbyhole with details
		_, err = vault.WriteToCubbyhole(
			"requests/" + hash,
			map[string]interface{}{
				"Policy":    policy,
				"Current":   policyOld,
				"New":       policyNew,
				"Requester": self.Data["display_name"].(string),
				"Required":  status.Required,
				"Progress":  0,
			})
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

		// fetch all current unseals from cubbyhole
		// under NO circumstances should the result be sent back to the client
		// this part assumes that the unseals key value is well-constructed
		unseals := []string{}
		if request.Progress > 0 {
			resp, err := vault.ReadFromCubbyhole("unseal/" + hash)
			if err != nil {
				return logError(c, err.Error(), "Could not check root generation status")
			}
			unseals = strings.Split(resp.Data["unseals"].(string), ";")
		}

		// add the provided new unseal token
		unseals = append(unseals, unsealKey)

		// if not enough unseals yet, store it and return progress
		if len(unseals) < request.Required {
			// delimit unseal tokens by semicolon
			_, err = vault.WriteToCubbyhole("unseal/" + hash,
				map[string]interface{}{
					"unseals": strings.Trim(strings.Join(strings.Fields(fmt.Sprint(unseals)), ";"), "[]"),
				})
			if err != nil {
				return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
			}

			// store progress in request too
			_, err = vault.WriteToCubbyhole(
				"requests/" + hash,
				map[string]interface{}{
					"Policy":    request.Policy,
					"Current":   request.Current,
					"New":       request.New,
					"Requester": request.Requester,
					"Required":  request.Required,
					"Progress":  len(unseals),
				})
			if err != nil {
				return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
			}

			// return progress
			return c.JSON(http.StatusOK, H{
				"progress": len(unseals),
			})
		}

		// if we got here, it means there are enough unseals to attempt root generation
		otp := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16))
		status, err = vault.GenerateRootInit(otp)
		if err != nil {
			return logError(c, err.Error(), "Another root generation may be in progress")
		}

		if status.EncodedRootToken == "" {
			for _, s := range(unseals) {
				status, err = vault.GenerateRootUpdate(s, status.Nonce)

				// an error likely means one of the unseals was not valid
				if err != nil {
					// delete root generation process
					if err := vault.GenerateRootCancel(); err != nil {
						return logError(c, err.Error(), "At least one unseal key was invalid. Could not revert root generation!")
					}

					// reset progress in request and cubbyhole
					if _, err := vault.WriteToCubbyhole("unseal/" + hash,
						map[string]interface{}{
							"unseals": "",
						}); err != nil {
							return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
						}
					if _, err := vault.WriteToCubbyhole(
						"requests/" + hash,
						map[string]interface{}{
							"Policy":    request.Policy,
							"Current":   request.Current,
							"New":       request.New,
							"Requester": request.Requester,
							"Required":  request.Required,
							"Progress":  0,
						}); err != nil {
							return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
						}

					// inform user that request unseals have been reset
					return c.JSON(http.StatusBadRequest, H{
						"error": "At least one unseal key was invalid. Progress has been reset.",
					})
				}

			}
		}

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
		defer vault.DeleteFromCubbyhole("unseals/" + hash)
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
