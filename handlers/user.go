package handlers

import (
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo"
)

func GetUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		result, err := auth.ListUsers(c.QueryParam("type"))
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// verify form data
		var deleteTarget = &vault.AuthInfo{}
		if err := c.Bind(deleteTarget); err != nil {
			return logError(c, err.Error(), "Invalid format")
		}
		if deleteTarget.Type == "" || deleteTarget.ID == "" {
			return logError(c, "Received empty delete request", "Invalid format")
		}

		// fetch auth from cookie
		getSession(c, auth)

		// delete user
		if err := auth.DeleteUser(deleteTarget.Type, deleteTarget.ID); err != nil {
			return logError(c, err.Error(), "Deletion error")
		}

		return c.JSON(http.StatusOK, H{
			"result": "User deleted successfully",
		})
	}
}
