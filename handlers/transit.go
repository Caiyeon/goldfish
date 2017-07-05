package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

func TransitInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// ensure token can lookup self before exposing transit key name
		if _, err := auth.Client(); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Invalid vault token",
			})
		}

		conf := vault.GetConfig()
		c.Response().Writer.Header().Set("UserTransitKey", conf.UserTransitKey)
		return c.JSON(http.StatusOK, H{
			"status": "fetched",
		})
	}
}

func EncryptString() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
