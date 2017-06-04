package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

func WrapHandler() echo.HandlerFunc {
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

		wrapttl := c.FormValue("wrapttl")
		if wrapttl == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "wrapttl cannot be 0",
			})
		}

		data := c.FormValue("data")

		// fetch results
		wrappingToken, err := auth.WrapData(wrapttl, data)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": wrappingToken,
		})
	}
}

func UnwrapHandler() echo.HandlerFunc {
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

		wrappingToken := c.FormValue("wrappingToken")
		if wrappingToken == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Wrapping token cannot be empty",
			})
		}

		// fetch results
		data, err := auth.UnwrapData(wrappingToken)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": data,
		})
	}
}
