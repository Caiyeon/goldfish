package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
	"github.com/gorilla/csrf"
	vaultapi "github.com/hashicorp/vault/api"
)

func GetMounts() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		mounts, err := auth.ListMounts()
		if err != nil {
			return logError(c, err.Error(), "Unauthorized")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": mounts,
		})
	}
}

func GetMount() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		result, err := auth.GetMount(c.Param("mountname"))
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func ConfigMount() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		var config *vaultapi.MountConfigInput
		if err := c.Bind(&config); err != nil {
			return logError(c, err.Error(), "Invalid format")
		}

		// fetch results
		err := auth.TuneMount(c.Param("mountname"), *config)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": "ok",
		})
	}
}