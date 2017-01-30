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
		conf := new(vaultConfig)
		if err := c.Bind(conf); err != nil {
			log.Println(err)
			return err
		}
		if conf.Addr == "" || conf.Token == "" {
			log.Println("Invalid authentication")
			return errors.New("Invalid authentication")
		}

		// check authentication
		l, err := lukko.NewLukko(conf.Addr, conf.Token)
		if err != nil {
			log.Println(err)
			return nil
		}
		defer l.Close()

		if err = l.CheckAuth(); err != nil {
			log.Println(conf)
			log.Println(err)
			return nil
		}

		// store items in session
		session, err := store.Get(c.Request(), "session-id")
		if err != nil {
			log.Println(err)
			return err
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
		result, err := l.ListAuth("token")
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, H{
			"tokens": result,
		})
	}
}