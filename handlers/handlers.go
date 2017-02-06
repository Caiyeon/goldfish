package handlers

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/labstack/echo"
)

// for returning JSON bodies
type H map[string]interface{}

// for storing ciphers of user credentials
var scookie = &securecookie.SecureCookie{}

// for authenticating this web server with vault
var vaultAddress = ""
var vaultToken = ""
var vaultClient *vaultapi.Client

// for binding login info
type AuthInfo struct {
	Type string `json:"Type" form:"Type" query:"Type"`
	ID   string `json:"ID" form:"ID" query:"ID"`
}

func init() {
	gob.Register(&AuthInfo{})

	// setup cookie encryption keys
	hashKey := securecookie.GenerateRandomKey(64)
	blockKey := securecookie.GenerateRandomKey(32)
	if hashKey == nil || blockKey == nil {
		panic("Failed to generate random hashkey")
	}
	scookie = securecookie.New(hashKey, blockKey)
	scookie = scookie.MaxAge(1800)
	if scookie == nil {
		panic("Failed to initialize gorilla/securecookie")
	}

	// read vault token and addr for web server
	// to do: change token to approle
	flag.StringVar(&vaultAddress, "addr", "http://127.0.0.1:8200", "Vault address")
	flag.StringVar(&vaultToken, "token", "", "Vault token")
	flag.Parse()
	if vaultAddress == "" || vaultToken == "" {
		panic("Invalid vault credentials")
	}

	// set up web server's vault client
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	client.SetAddress(vaultAddress)
	client.SetToken(vaultToken)
	if _, err = client.Auth().Token().LookupSelf(); err != nil {
		panic(err)
	}
	vaultClient = client
}

// error handler should print error in server log
// but return no unnecessary server info to client
func handleError(c echo.Context, logstring string, responsestring string) error {
	log.Println("[ERROR]:", logstring)
	return c.JSON(http.StatusInternalServerError, H{
		"error": responsestring,
	})
}

func FetchCSRF() echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))
		return c.JSON(http.StatusOK, H{
			"status": "fetched",
		})
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := new(AuthInfo)
		defer auth.clear()

		// read form data
		if err := c.Bind(auth); err != nil {
			return handleError(c, err.Error(), "Invalid format")
		}
		if auth.Type == "" || auth.ID == "" {
			return handleError(c, "Empty authentication", "Empty authentication")
		}

		// verify auth details
		if _, err := auth.client(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// encrypt auth.ID with vault's transit backend
		if err := auth.encrypt(); err != nil {
			return handleError(c, err.Error(), "Internal error")
		}

		// store auth.Type and auth.ID (now a cipher) in cookie
		if encoded, err := scookie.Encode("auth", auth); err == nil {
			cookie := &http.Cookie{
				Name:  "auth",
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(c.Response().Writer, cookie)
		} else {
			return handleError(c, err.Error(),
				"Please clear site-related cookie and storage",
			)
		}

		// success
		return c.JSON(http.StatusOK, H{
			"status": "Logged in",
		})
	}
}

func GetUsers() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &AuthInfo{}
		defer auth.clear()

		// fetch auth from cookie
		if cookie, err := c.Request().Cookie("auth"); err == nil {
			if err = scookie.Decode("auth", cookie.Value, &auth); err != nil {
				return handleError(c, err.Error(), "Please clear cookies and login again")
			}
		} else {
			return handleError(c, err.Error(), "Please clear cookies and login again")
		}

		// decode auth's ID with vault transit backend
		if err := auth.decrypt(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// verify auth details
		if _, err := auth.client(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// fetch results
		result, err := auth.listusers(c.QueryParam("type"))
		if err != nil {
			return handleError(c, err.Error(), "Internal error")
		}

		// give a CSRF token in case a delete request is sent later
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		// return result
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &AuthInfo{}
		defer auth.clear()

		// read form data
		var deleteTarget = &AuthInfo{}
		if err := c.Bind(deleteTarget); err != nil {
			return handleError(c, err.Error(), "Invalid format")
		}
		if deleteTarget.Type == "" || deleteTarget.ID == "" {
			return handleError(c, "Received empty delete request", "Invalid format")
		}

		// fetch auth from cookie
		if cookie, err := c.Request().Cookie("auth"); err == nil {
			if err = scookie.Decode("auth", cookie.Value, &auth); err != nil {
				return handleError(c, err.Error(), "Please clear cookies and login again")
			}
		} else {
			return handleError(c, err.Error(), "Please clear cookies and login again")
		}

		// decode auth's ID with vault transit backend
		if err := auth.decrypt(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// verify auth details
		if _, err := auth.client(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// delete user
		if err := auth.deleteuser(deleteTarget.Type, deleteTarget.ID); err != nil {
			return handleError(c, err.Error(), "Deletion error")
		}

		return c.JSON(http.StatusOK, H{
			"result": "User deleted successfully",
		})
	}
}

func GetPolicies() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &AuthInfo{}
		defer auth.clear()

		// fetch auth from cookie
		if cookie, err := c.Request().Cookie("auth"); err == nil {
			if err = scookie.Decode("auth", cookie.Value, &auth); err != nil {
				return handleError(c, err.Error(), "Please clear cookies and login again")
			}
		} else {
			return handleError(c, err.Error(), "Please clear cookies and login again")
		}

		// decode auth's ID with vault transit backend
		if err := auth.decrypt(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// verify auth details
		if _, err := auth.client(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// fetch results
		result, err := auth.listpolicies()
		if err != nil {
			return handleError(c, err.Error(), "Internal error")
		}

		// give a CSRF token in case a delete request is sent later
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		// return result
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func GetPolicy() echo.HandlerFunc {
	return func(c echo.Context) error {
		var auth = &AuthInfo{}
		defer auth.clear()

		// fetch auth from cookie
		if cookie, err := c.Request().Cookie("auth"); err == nil {
			if err = scookie.Decode("auth", cookie.Value, &auth); err != nil {
				return handleError(c, err.Error(), "Please clear cookies and login again")
			}
		} else {
			return handleError(c, err.Error(), "Please clear cookies and login again")
		}

		// decode auth's ID with vault transit backend
		if err := auth.decrypt(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// verify auth details
		if _, err := auth.client(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// fetch results
		result, err := auth.getpolicy(c.Param("policyname"))
		if err != nil {
			return handleError(c, err.Error(), "Internal error")
		}

		// give a CSRF token in case a delete request is sent later
		c.Response().Writer.Header().Set("X-CSRF-Token", csrf.Token(c.Request()))

		// return result
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}
