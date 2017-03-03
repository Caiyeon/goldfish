package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

func EncryptString() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		var plaintext = &StringBind{}
		if err := c.Bind(plaintext); err != nil {
			return logError(c, err.Error(), "Invalid format")
		}

		// fetch results
		cipher, err := auth.EncryptTransit(plaintext.Str)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": cipher,
		})
	}
}

func DecryptString() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		var cipher = &StringBind{}
		if err := c.Bind(cipher); err != nil {
			return logError(c, err.Error(), "Invalid format")
		}

		// fetch results
		plaintext, err := auth.DecryptTransit(cipher.Str)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": plaintext,
		})
	}
}