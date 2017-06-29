package handlers

import (
	"net/http"

	"github.com/gorilla/csrf"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/labstack/echo"
)

func GetMounts() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		mounts, err := auth.ListMounts()
		if err != nil {
			return parseError(c, err)
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": mounts,
		})
	}
}

func GetMount() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch results
		result, err := auth.GetMount(c.Param("mountname"))
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func ConfigMount() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		var config *vaultapi.MountConfigInput
		if err := c.Bind(&config); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Invalid config format",
			})
		}

		// fetch results
		err := auth.TuneMount(c.Param("mountname"), *config)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "ok",
		})
	}
}
