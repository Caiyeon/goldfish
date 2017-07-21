package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func GetUserpassUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch results
		results, err := auth.ListUserpassUsers()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": results,
		})
	}
}

func DeleteUserpassUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// verify form data
		username := c.QueryParam("username")
		if username == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "username parameter is required",
			})
		}
		if _, err := auth.DeleteRaw("auth/userpass/users/" + username); err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "User deleted successfully",
		})
	}
}
