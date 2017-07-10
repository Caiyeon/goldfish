package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func GetBulletins() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		bulletins, err := auth.GetBulletins()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": bulletins,
		})
	}
}
