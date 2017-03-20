package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo"
)

func GetSecrets() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		path := c.QueryParam("path")
		if path == "" {
			return logError(c, "Empty path parameter in getting secrets", "Invalid parameter")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		if path[len(path)-1:] == "/" {
			// listing a directory
			if result, err := auth.ListSecret(path); err != nil {
				return logError(c, err.Error(), "Internal error")
			} else {
				return c.JSON(http.StatusOK, H{
					"result": result,
				})
			}
		} else {
			// reading a specific secret's key value pairs
			if result, err := auth.ReadSecret(path); err != nil {
				return logError(c, err.Error(), "Internal error")
			} else {
				return c.JSON(http.StatusOK, H{
					"result": result,
				})
			}
		}
	}
}

func PostSecrets() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		path := c.QueryParam("path")
		body := c.FormValue("body")

		if path == "" || body == "" {
			return logError(c, "Empty path or body", "Path and body cannot be empty")
		}

		if path[len(path)-1:] == "/" {
			return logError(c, "Invalid path", "Path must not end in '/'")
		}

		resp, err := auth.WriteSecret(path, body)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": resp,
		})
	}
}
