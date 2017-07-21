package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func GetApproleRoles() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch results
		results, err := auth.ListApproleRoles()
		if err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": results,
		})
	}
}

func DeleteApproleRole() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header or cookie
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// verify form data
		role := c.QueryParam("role")
		if role == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "role parameter is required",
			})
		}
		if _, err := auth.DeleteRaw("auth/approle/role/" + role); err != nil {
			return parseError(c, err)
		}

		return c.JSON(http.StatusOK, H{
			"result": "Approle role deleted successfully",
		})
	}
}
