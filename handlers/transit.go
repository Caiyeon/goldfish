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

		plaintext := c.FormValue("plaintext")
		if plaintext == "" {
			return logError(c, "Empty plaintext provided", "Plaintext is empty")
		}

		// fetch results
		cipher, err := auth.EncryptTransit(plaintext)
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

		cipher := c.FormValue("cipher")
		if cipher == "" {
			return logError(c, "Empty cipher provided", "Cipher is empty")
		}

		// fetch results
		plaintext, err := auth.DecryptTransit(cipher)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": plaintext,
		})
	}
}