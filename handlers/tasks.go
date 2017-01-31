package handlers

import (
	"encoding/gob"
	"net/http"
	"errors"
	"log"

	"github.com/labstack/echo"
	"github.com/gorilla/sessions"
	"github.com/caiyeon/lukko"
)

// for returning JSON bodies
type H map[string]interface{}

// for storing session data
var store = sessions.NewCookieStore([]byte("to-be-made-secret"))

// for binding login info
type vaultConfig struct {
	Addr  string `json:"addr" form:"addr" query:"addr"`
	Token string `json:"token" form:"token" query:"addr"`
}

func init() {
	gob.Register(&vaultConfig{})
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

		// store items in session
		session, err := store.Get(c.Request(), "session-id")
		if err != nil {
			returnError(err.Error())
		}
		session.Values["vaultConfig"] = conf
		session.Save(c.Request(), c.Response().Writer)

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
			return errors.New("type must be non-empty")
		}

		// check session for authentication status
		session, err := store.Get(c.Request(), "session-id")
		if err != nil {
			return err
		}

		// extract config from session cookie
		raw := session.Values["vaultConfig"]
		var conf = &vaultConfig{}
		conf, ok := raw.(*vaultConfig)
		if !ok {
			return errors.New("Failed to read session cookie")
		}

		// check authentication
		l, err := lukko.NewLukko(conf.Addr, conf.Token)
		if err != nil {
			return nil
		}
		if err = l.CheckAuth(); err != nil {
			return nil
		}
		defer l.Close()

		// return all tokens
		result, err := l.ListAuth(authtype)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, H{
			"result": result,
		})
	}
}