package handlers

import (
	"encoding/gob"
	"net/http"
	"errors"
	"log"

	"github.com/labstack/echo"
	// uuid "github.com/hashicorp/go-uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/csrf"
	"github.com/caiyeon/lukko"
)

// for returning JSON bodies
type H map[string]interface{}

// for storing session data
// var store = sessions.NewCookieStore([]byte("to-be-made-secret"))
var scookie = &securecookie.SecureCookie{}

// for binding login info
type vaultConfig struct {
	Addr  string `json:"addr" form:"addr" query:"addr"`
	Token string `json:"token" form:"token" query:"addr"`
}

func init() {
	gob.Register(&vaultConfig{})

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
		// error handler should print error in server log
		// but return no unnecessary server info to client
		returnError := func(s string) error {
			log.Println("[ERROR]: Login:", s)
			return c.JSON(http.StatusInternalServerError, H{
				"status": "Login failed",
			})
		}

		// read form data
		conf := new(vaultConfig)
		if err := c.Bind(conf); err != nil {
			returnError(err.Error())
		}
		if conf.Addr == "" || conf.Token == "" {
			returnError("Invalid authentication")
		}

		// check authentication
		l, err := lukko.NewLukko(conf.Addr, conf.Token)
		if err != nil {
			returnError(err.Error())
		}
		defer l.Close()
		if err = l.CheckAuth(); err != nil {
			returnError(err.Error())
		}

		if encoded, err := scookie.Encode("auth", conf); err == nil {
			cookie := &http.Cookie{
				Name: "auth",
				Value: encoded,
				Path: "/",
			}
			http.SetCookie(c.Response().Writer, cookie)
		} else {
			returnError(err.Error())
		}

		// return display name
		return c.JSON(http.StatusOK, H{
			"status": "Logged in",
		})
	}
}

func Users() echo.HandlerFunc {
	return func(c echo.Context) error {
		authtype := c.QueryParam("type")
		if authtype == "" {
			log.Println("type empty")
			return errors.New("type must be non-empty")
		}

		var conf = &vaultConfig{}
		if cookie, err := c.Request().Cookie("auth"); err == nil {
			if err = scookie.Decode("auth", cookie.Value, &conf); err != nil {
				log.Println(err.Error())
			}
		} else {
			log.Println(err.Error())
		}

		// check authentication
		l, err := lukko.NewLukko(conf.Addr, conf.Token)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		if err = l.CheckAuth(); err != nil {
			log.Println(err.Error())
			log.Println("token:", conf.Token)
			return nil
		}
		defer l.Close()

		// return all tokens
		result, err := l.ListAuth(authtype)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}

func DeleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		authtype := c.QueryParam("type")
		id := c.QueryParam("id")
		if authtype == "" || id == "" {
			log.Println("type must be non-empty")
			return errors.New("type must be non-empty")
		}

		var path string
		switch authtype {
		case "token":
			path = "/auth/token/revoke-accessor/" + id
		case "userpass":
			path = "/auth/userpass/users/" + id
		default:
			log.Println("Authtype not supported")
			return errors.New("Authtype not supported")
		}

		var conf = &vaultConfig{}
		if cookie, err := c.Request().Cookie("auth"); err != nil {
			if err = scookie.Decode("auth", cookie.Value, &conf); err != nil {
				log.Println(err.Error())
			}
		}

		// check authentication
		l, err := lukko.NewLukko(conf.Addr, conf.Token)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		if err = l.CheckAuth(); err != nil {
			log.Println(err.Error())
			return nil
		}
		defer l.Close()

		resp, err := l.Delete(authtype, path)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		return c.JSON(http.StatusOK, H{
			"result": resp,
		})
	}
}