package handlers

import (
	"log"
	"net/http"

	"github.com/caiyeon/goldfish/vault"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
)

// for returning JSON bodies
type H map[string]interface{}

// for storing ciphers of user credentials
var scookie = &securecookie.SecureCookie{}

// constraint by vue-resource. Will not need if/when switched to axios
type StringBind struct {
	Str string `json:"Str" form:"Str" query:"Str"`
}

func init() {
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
}

func logError(c echo.Context, logstring string, responsestring string) error {
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

func VaultHealth() echo.HandlerFunc {
	return func(c echo.Context) error {
		resp, err := vault.VaultHealth()
		if err != nil {
			return logError(c, "Failed to reach vault health endpoint", "Internal error")
		}
		return c.JSON(http.StatusOK, H{
			"result": string(resp),
		})
	}
}

func Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := new(vault.AuthInfo)
		defer auth.Clear()

		// read form data
		if err := c.Bind(auth); err != nil {
			return logError(c, err.Error(), "Invalid format")
		}
		if auth.Type == "" || auth.ID == "" {
			return logError(c, "Empty authentication", "Empty authentication")
		}

		// verify auth details
		if _, err := auth.Client(); err != nil {
			return logError(c, err.Error(), "Invalid authentication")
		}

		// encrypt auth.ID with vault's transit backend
		if err := auth.EncryptAuth(); err != nil {
			return logError(c, err.Error(), "Internal error")
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
			return logError(c, err.Error(),
				"Please clear site-related cookie and storage",
			)
		}

		// To do: return list of policies
		return c.JSON(http.StatusOK, H{
			"status": "Logged in",
		})
	}
}

func getSession(c echo.Context, auth *vault.AuthInfo) error {
	// fetch auth from cookie
	if cookie, err := c.Request().Cookie("auth"); err == nil {
		if err = scookie.Decode("auth", cookie.Value, &auth); err != nil {
			return logError(c, err.Error(), "Please clear cookies and login again")
		}
	} else {
		return logError(c, err.Error(), "Please clear cookies and login again")
	}

	// decode auth's ID with vault transit backend
	if err := auth.DecryptAuth(); err != nil {
		return logError(c, err.Error(), "Invalid authentication")
	}
	return nil
}