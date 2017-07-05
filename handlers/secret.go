package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

func GetSecrets() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		path := c.QueryParam("path")
		if path == "" {
			conf := vault.GetConfig()
			path = conf.DefaultSecretPath
		}

		if path == "" || path[len(path)-1:] == "/" {
			// listing a directory
			if result, err := auth.ListSecret(path); err != nil {
				return parseError(c, err)
			} else {
				return c.JSON(http.StatusOK, H{
					"result": result,
					"path":   path,
				})
			}
		} else {
			// reading a specific secret's key value pairs
			if result, err := auth.ReadSecret(path); err != nil {
				return parseError(c, err)
			} else {
				return c.JSON(http.StatusOK, H{
					"result": result,
					"path":   path,
				})
			}
		}
	}
}

func PostSecrets() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		path := c.QueryParam("path")
		body := c.FormValue("body")

		if path == "" || body == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Path and body must not be empty",
			})
		}

		if path[len(path)-1:] == "/" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Path must not end in '/'",
			})
		}

		resp, err := auth.WriteSecret(path, body)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": resp,
		})
	}
}

func DeleteSecrets() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		_, err := auth.DeleteSecret(c.QueryParam("path"))
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "success",
		})
	}
}
