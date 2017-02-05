package handlers

import (
	"encoding/gob"
	"net/http"
	// "errors"
	"log"
	"flag"

	"github.com/labstack/echo"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/csrf"
	// "github.com/caiyeon/lukko"
	vaultapi "github.com/hashicorp/vault/api"
)

// for returning JSON bodies
type H map[string]interface{}

var scookie = &securecookie.SecureCookie{}

// for authenticating this web server with vault
var vaultAddress = ""
var vaultToken = ""
var vaultClient *vaultapi.Client

// for binding login info
type AuthInfo struct {
	Type  string `json:"Type" form:"Type" query:"Type"`
	ID    string `json:"ID" form:"ID" query:"ID"`
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
	scookie = scookie.MaxAge(300)
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
	if setupClient() != nil {
		panic("Failed to setup vault client for web server")
	}
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
		// read form data
		auth := new(AuthInfo)
		if err := c.Bind(auth); err != nil {
			return handleError(c, err.Error(), "Invalid format")
		}
		if auth.Type == "" || auth.ID == "" {
			return handleError(c, "Empty authentication", "Empty authentication")
		}

		// verify auth details are ok
		if err := auth.check(); err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// store auth in vault cubbyhole
		path, err := auth.store()
		if err != nil {
			return handleError(c, err.Error(), "Invalid authentication")
		}

		// store cubbyhole's path in securecookie
		if encoded, err := scookie.Encode("auth", path); err == nil {
			cookie := &http.Cookie{
				Name: "auth",
				Value: encoded,
				Path: "/",
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

// func Users() echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		authtype := c.QueryParam("type")
// 		if authtype == "" {
// 			log.Println("type empty")
// 			return errors.New("type must be non-empty")
// 		}

// 		var conf = &vaultConfig{}
// 		if cookie, err := c.Request().Cookie("auth"); err == nil {
// 			if err = scookie.Decode("auth", cookie.Value, &conf); err != nil {
// 				log.Println(err.Error())
// 			}
// 		} else {
// 			log.Println(err.Error())
// 		}

// 		// check authentication
// 		l, err := lukko.NewLukko(conf.Addr, conf.Token)
// 		if err != nil {
// 			log.Println(err.Error())
// 			return nil
// 		}
// 		if err = l.CheckAuth(); err != nil {
// 			log.Println(err.Error())
// 			log.Println("token:", conf.Token)
// 			return nil
// 		}
// 		defer l.Close()

// 		// return all tokens
// 		result, err := l.ListAuth(authtype)
// 		if err != nil {
// 			log.Println(err.Error())
// 			return err
// 		}
// 		return c.JSON(http.StatusOK, H{
// 			"result": result,
// 		})
// 	}
// }

// func DeleteUser() echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		authtype := c.QueryParam("type")
// 		id := c.QueryParam("id")
// 		if authtype == "" || id == "" {
// 			log.Println("type must be non-empty")
// 			return errors.New("type must be non-empty")
// 		}

// 		var path string
// 		switch authtype {
// 		case "token":
// 			path = "/auth/token/revoke-accessor/" + id
// 		case "userpass":
// 			path = "/auth/userpass/users/" + id
// 		default:
// 			log.Println("Authtype not supported")
// 			return errors.New("Authtype not supported")
// 		}

// 		var conf = &vaultConfig{}
// 		if cookie, err := c.Request().Cookie("auth"); err != nil {
// 			if err = scookie.Decode("auth", cookie.Value, &conf); err != nil {
// 				log.Println(err.Error())
// 			}
// 		}

// 		// check authentication
// 		l, err := lukko.NewLukko(conf.Addr, conf.Token)
// 		if err != nil {
// 			log.Println(err.Error())
// 			return nil
// 		}
// 		if err = l.CheckAuth(); err != nil {
// 			log.Println(err.Error())
// 			return nil
// 		}
// 		defer l.Close()

// 		resp, err := l.Delete(authtype, path)
// 		if err != nil {
// 			log.Println(err.Error())
// 			return err
// 		}

// 		return c.JSON(http.StatusOK, H{
// 			"result": resp,
// 		})
// 	}
// }