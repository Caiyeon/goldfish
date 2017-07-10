package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func WrapHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
