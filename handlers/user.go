package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/hashicorp/vault/api"
	"github.com/labstack/echo"
)

func GetUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		var offset int
		var err error
		if c.QueryParam("offset") == "" {
			offset = 0
		} else {
			offset, err = strconv.Atoi(c.QueryParam("offset"))
			if err != nil {
				return logError(c, err.Error(), "Internal error")
			}
		}

		// fetch results
		result, err := auth.ListUsers(c.QueryParam("type"), offset)
		if err != nil {
			return logError(c, err.Error(), "Internal error")
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func GetTokenCount() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		// fetch results
		result, err := auth.GetTokenCount()
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

func CreateUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		var resp *api.Secret
		switch c.QueryParam("type") {
		case "":
			return logError(c, "Received empty user creation type", "Creation type cannot be empty")

		case "token":
			var request = &api.TokenCreateRequest{}
			err := c.Bind(request)
			if err != nil {
				return logError(c, err.Error(), "Invalid format")
			}

			resp, err = auth.CreateToken(request)
			if err != nil {
				return logError(c, err.Error(), "Could not create token")
			}

		default:
			return logError(c, "Received unknown creation type", "Unsupported creation type")
		}

		return c.JSON(http.StatusOK, H{
			"result": resp,
		})
	}
}

func ListRoles() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		result, err := auth.ListRoles()
		if err != nil {
			log.Println("[ERROR]:", err.Error())
			return c.JSON(http.StatusForbidden, H{
				"error": "Could not list roles",
			})
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func GetRole() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &vault.AuthInfo{}
		defer auth.Clear()

		// fetch auth from cookie
		getSession(c, auth)

		result, err := auth.GetRole(c.QueryParam("rolename"))
		if err != nil {
			return logError(c, err.Error(), "Could not read role")
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}
