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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		bulletins, err := auth.GetBulletins()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": bulletins,
		})
	}
}
