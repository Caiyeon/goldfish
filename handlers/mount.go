package handlers

import (
	"net/http"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/labstack/echo"
)

func GetMount() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// if no mount is specified, list all mounts
		if mount := c.QueryParam("mount"); mount == "" {
			result, err := auth.ListMounts()
			if err != nil {
				return parseError(c, err)
			}
			return c.JSON(http.StatusOK, H{
				"result": result,
			})
		} else {
			result, err := auth.GetMount(mount)
			if err != nil {
				return parseError(c, err)
			}
			return c.JSON(http.StatusOK, H{
				"result": result,
			})
		}
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
		err := auth.TuneMount(c.QueryParam("mount"), *config)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "ok",
		})
	}
}
