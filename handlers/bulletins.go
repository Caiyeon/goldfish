package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

func GetBulletins() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		bulletins, err := auth.GetBulletins()
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		return c.JSON(http.StatusOK, H{
			"result": bulletins,
		})
	}
}
