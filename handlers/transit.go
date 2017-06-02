package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo"
)

func TransitInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		// ensure user has default policy before giving transit key name
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		conf := vault.GetConfig()
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))
		c.Response().Writer.Header().Set("UserTransitKey", conf.UserTransitKey)
		return c.JSON(http.StatusOK, H{
			"status": "fetched",
		})
	}
}

func EncryptString() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		plaintext := c.FormValue("plaintext")
		if plaintext == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Plaintext must not be empty",
			})
		}

		// fetch results
		cipher, err := auth.EncryptTransit(c.FormValue("key"), plaintext)
		if err != nil {
			return parseError(c, err)
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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		cipher := c.FormValue("cipher")
		if cipher == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Cipher must not be empty",
			})
		}

		// fetch results
		plaintext, err := auth.DecryptTransit(c.FormValue("key"), cipher)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": plaintext,
		})
	}
}
