package main

import (
	"github.com/caiyeon/goldfish/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
)

func main() {
	// create a new instance of Echo
	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// routing
	e.Static("/", "public")

	e.GET("/login/csrf", handlers.FetchCSRF())
	e.POST("/login", handlers.Login())
	// e.GET("/users", handlers.Users())
	// e.DELETE("/users", handlers.DeleteUser())

	e.Use(echo.WrapMiddleware(
		csrf.Protect(
			[]byte(securecookie.GenerateRandomKey(32)),
			csrf.Secure(false),
		)))

	// Start a web server
	e.Logger.Fatal(e.Start(":8000"))
}