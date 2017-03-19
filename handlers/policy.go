package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo"
)

func ListPolicies() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		result, err := auth.ListPolicies()
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func GetPolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		result, err := auth.GetPolicy(c.Param("policyname"))
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
		if err := auth.DeletePolicy(c.Param("policyname")); err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": "Policy deleted",
		})
	}
}
