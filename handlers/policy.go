package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo"
)

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
