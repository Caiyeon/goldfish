package handlers

import (
	"net/http"
	"strconv"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/hashicorp/hcl"
	"github.com/labstack/echo"
	"github.com/mitchellh/hashstructure"
)

type PolicyRequest struct {
	Policy    string
	Current   string
	New       string
	Requester string
	Required  int
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
// Requires goldfish server to have write access to /sys/generate-root/*
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
			return logError(c, err.Error(), "No changes detected")
		}

		// collect non-dangerous identifying data on requester
		self, err := auth.LookupSelf()
		if err != nil {
			return logError(c, err.Error(), "Failed to perform lookupself on requester token")
		}

		// get number of unseal keys required to generate root token
		// this requires goldfish server token to be able to read /sys/generate-root/attempt
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
		}

		// hash structure
		hash_uint64, err := hashstructure.Hash(request, nil)
		if err != nil {
			return logError(c, err.Error(), "Could not hash request. Unsafe; aborting.")
		}
		hash := strconv.FormatUint(hash_uint64, 16)

		// write to cubbyhole with details
		_, err = vault.WriteToCubbyhole(
			hash,
			map[string]interface{}{
				"Policy":    policy,
				"Current":   policyOld,
				"New":       policyNew,
				"Requester": self.Data["display_name"].(string),
				"Required":  status.Required,
			})
		if err != nil {
			return logError(c, err.Error(), "Could not save to cubbyhole. Unsafe; aborting.")
		}

		// return hash
		return c.JSON(http.StatusOK, H{
			"result": hash,
		})
	}
}
