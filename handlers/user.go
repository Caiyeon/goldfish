package handlers

import (
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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		var offset int
		var err error
		if c.QueryParam("offset") == "" {
			offset = 0
		} else {
			offset, err = strconv.Atoi(c.QueryParam("offset"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, H{
					"error": "Offset is not an integer",
				})
			}
		}

		// fetch results
		result, err := auth.ListUsers(c.QueryParam("type"), offset)
		if err != nil {
			return parseError(c, err)
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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		// fetch results
		result, err := auth.GetTokenCount()
		if err != nil {
			return parseError(c, err)
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
			return c.JSON(http.StatusBadRequest, H{
				"error": "Invalid format for deletion target",
			})
		}
		if deleteTarget.Type == "" || deleteTarget.ID == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Deletion target cannot be empty",
			})
		}

		// fetch auth from cookie
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		// delete user
		if err := auth.DeleteUser(deleteTarget.Type, deleteTarget.ID); err != nil {
			return parseError(c, err)
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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		var resp *api.Secret
		switch c.QueryParam("type") {
		case "":
			return c.JSON(http.StatusBadRequest, H{
				"error": "User creation type cannot be empty",
			})

		case "token":
			var request = &api.TokenCreateRequest{}
			err := c.Bind(request)
			if err != nil {
				return c.JSON(http.StatusBadRequest, H{
					"error": "Invalid token creation format",
				})
			}

			resp, err = auth.CreateToken(request, c.QueryParam("wrap-ttl"))
			if err != nil {
				return parseError(c, err)
			}

		default:
			return c.JSON(http.StatusBadRequest, H{
				"error": "User creation type not supported",
			})
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
		if err := getSession(c, auth); err != nil {
			return c.JSON(http.StatusForbidden, H{
				"error": "Please login first",
			})
		}
		if err := auth.DecryptAuth(); err != nil {
			return parseError(c, err)
		}

		// check if user has access to roles
		capabilities, err := auth.CapabilitiesSelf("auth/token/roles")
		if err != nil {
			return parseError(c, err)
		}
		capabilities2, err := auth.CapabilitiesSelf("auth/token/roles/")
		if err != nil {
			return parseError(c, err)
		}

		for _, capability := range append(capabilities, capabilities2...) {
			// if user can list or is root, return list of roles
			if capability == "list" || capability == "root" {
				result, err := auth.ListRoles()
				if err != nil {
					return parseError(c, err)
				}

				return c.JSON(http.StatusOK, H{
					"result": result,
				})
			}
		}

		// if we got here, it means user is authenticated against vault,
		// but has no list capability on roles
		return c.JSON(http.StatusForbidden, H{
			"error": "User lacks capability to list roles",
		})
	}
}

func GetRole() echo.HandlerFunc {
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

		result, err := auth.GetRole(c.QueryParam("rolename"))
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}
