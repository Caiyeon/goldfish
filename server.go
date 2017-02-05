package main

import (
	"github.com/caiyeon/goldfish/handlers"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(echo.WrapMiddleware(
		csrf.Protect(
			// Generate a new encryption key each launch
			[]byte(securecookie.GenerateRandomKey(32)),
			csrf.Secure(false),
		)))

	// routing
	e.Static("/", "public")
	e.GET("/login/csrf", handlers.FetchCSRF())
	e.POST("/login", handlers.Login())
	e.GET("/users", handlers.Users())
	e.DELETE("/users", handlers.DeleteUser())

	// start the server
	e.Logger.Fatal(e.Start(":8000"))
}
