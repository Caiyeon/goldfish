package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

// for returning JSON bodies
type H map[string]interface{}

// returns the http status code found in the error message
func parseError(c echo.Context, err error) error {
	// if error came from vault, relay it
	errCode := strings.Split(err.Error(), "Code:")
	errMsgs := strings.Split(err.Error(), "*")
	if len(errCode) > 1 && len(errMsgs) > 1 {
		code := 500
		fmt.Sscanf(errCode[1], "%d", &code)
		return c.JSON(code, H{
			"error": "Vault: " + errMsgs[1],
		})
	}

	// if error came from goldfish
	log.Println("[ERROR]: ", err.Error())
	return c.JSON(http.StatusInternalServerError, H{
		"error": err.Error(),
	})
}

func VaultHealth() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, err := vault.VaultHealth()
		if err != nil {
			return parseError(c, err)
		}
		return c.JSON(http.StatusOK, H{
			"result": string(resp),
		})
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := new(vault.AuthInfo)
		defer auth.Clear()

		// read form data
		if err := c.Bind(auth); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Invalid auth format",
			})
		}
		if auth.Type == "" || auth.ID == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Empty authentication",
			})
		}

		// verify auth details and create client access token
		data, err := auth.Login()
		if err != nil {
			return parseError(c, err)
		}

		// if goldfish is configured to use transit encryption
		if conf := vault.GetConfig(); conf.ServerTransitKey != "" {
			// encrypt auth.ID with vault's transit backend
			if err := auth.EncryptAuth(); err != nil {
				return c.JSON(http.StatusInternalServerError, H{
					"error": "Goldfish could not use transit key: " + err.Error(),
				})
			}
		}

		// return useful information to user
		return c.JSON(http.StatusOK, H{
			"status": "Logged in",
			"result": map[string]interface{}{
				"cipher":       auth.ID,
				"display_name": data["display_name"],
				"id":           data["id"],
				"meta":         data["meta"],
				"policies":     data["policies"],
				"renewable":    data["renewable"],
				"ttl":          data["ttl"],
			},
		})
	}
}

func RenewSelf() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// verify auth details and create client access token
		resp, err := auth.RenewSelf()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": map[string]interface{}{
				"meta":     resp.Auth.Metadata,
				"policies": resp.Auth.Policies,
				"ttl":      resp.Auth.LeaseDuration,
			},
		})
	}
}

// reads header as an encrypted
func getSession(c echo.Context) *vault.AuthInfo {
	var auth = &vault.AuthInfo{
		Type: "token",
	}

	// check headers first
	if auth.ID = c.Request().Header.Get("X-Vault-Token"); auth.ID == "" {
		c.JSON(http.StatusForbidden, H{
			"error": "Please login first",
		})
		return nil
	}

	// if header is transit encrypted, decrypt first
	if strings.HasPrefix(auth.ID, "vault:") {
		if err := auth.DecryptAuth(); err != nil {
			c.JSON(http.StatusForbidden, H{
				"error": "Cipher invalid. Please logout and login again",
			})
			return nil
		}
	}

	return auth
}
