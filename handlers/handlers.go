package handlers

import (
	"encoding/json"
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
			"result": resp,
		})
	}
}

func Health() echo.HandlerFunc {
	return func(c echo.Context) error {
		bootstrapped := vault.Bootstrapped()

		deployment_time_utc := ""
		if bootstrapped {
			// check server token
			resp, err := vault.LookupSelf()
			if err != nil {
				return parseError(c, err)
			}
			deployment_time_utc = string(resp["creation_time"].(json.Number))
		}

		// check transit encryption config
		transitEnabled := vault.GetConfig().ServerTransitKey != ""

		return c.JSON(http.StatusOK, H{
			"bootstrapped":        bootstrapped,
			"deployment_time_utc": deployment_time_utc,
			"transit_encryption":  transitEnabled,
		})
	}
}

func Bootstrap() echo.HandlerFunc {
	// scoped struct is fine, nothing else needs to know this
	type wrapstruct struct {
		Wrapping_token string
	}

	return func(c echo.Context) error {
		// re-bootstrapping may come in the future, but isn't supported for now
		if vault.Bootstrapped() {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Already bootstrapped",
			})
		}

		// bind body
		wrap := new(wrapstruct)
		if err := c.Bind(wrap); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Invalid format",
			})
		}
		if wrap.Wrapping_token == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Empty wrapping token",
			})
		}

		if err := vault.Bootstrap(wrap.Wrapping_token); err != nil {
			return c.JSON(http.StatusInternalServerError, H{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, H{
			"result": "success",
		})
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		// if vault wrapper is not initialized, errors for everyone!
		if !vault.Bootstrapped() {
			c.JSON(http.StatusForbidden, H{
				"error": "Goldfish is not initialized!",
			})
			return nil
		}

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

		// verify auth details and renew access token
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

func RevokeSelf() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// verify auth details and revoke self
		err := auth.RevokeSelf()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": map[string]interface{}{
				"status": "success",
			},
		})
	}
}

// constructs raw or decrypted authentication info
func getSession(c echo.Context) *vault.AuthInfo {
	// if vault wrapper is not initialized, errors for everyone!
	if !vault.Bootstrapped() {
		c.JSON(http.StatusForbidden, H{
			"error": "Goldfish is not initialized!",
		})
		return nil
	}

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
