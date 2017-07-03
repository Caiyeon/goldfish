package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/hashicorp/vault/api"
	"github.com/labstack/echo"
)

func GetUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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

func GetTokenAccessors() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch results
		result, err := auth.GetTokenAccessors()
		if err != nil {
			return parseError(c, err)
		}

		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func LookupTokenByAccessor() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// scoped struct. No other functions need to know this
		type body struct {
			Accessors string `json:"accessors"`
		}
		var b = &body{}

		// input can be in param or body (comma separated)
		var err error
		b.Accessors = c.QueryParam("accessors")
		if b.Accessors == "" {
			if err = c.Bind(&b); err == nil {
				if b.Accessors == "" {
					err = errors.New("Required key 'accessors' not found in body")
				}
			}
		}
		if err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Invalid body: " + err.Error(),
			})
		}

		// fetch results
		result, err := auth.LookupTokenByAccessor(b.Accessors)
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeleteTokenByAccessor() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		err := auth.DeleteTokenByAccessor(c.QueryParam("accessor"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, H{
			"result": "Token deleted successfully",
		})
	}
}

func DeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

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
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		result, err := auth.GetRole(c.QueryParam("rolename"))
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}
