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

	e.GET("/health", handlers.GetHealth())

	e.GET("/login/csrf", handlers.FetchCSRF())
	e.POST("/login", handlers.Login())

	e.GET("/users", handlers.GetUsers())
	e.DELETE("/users", handlers.DeleteUser())

	e.GET("/policies", handlers.GetPolicies())
	e.GET("/policies/:policyname", handlers.GetPolicy())
	e.DELETE("/policies/:policyname", handlers.DeletePolicy())

	e.GET("/transit", handlers.FetchCSRF())
	e.POST("/transit/encrypt", handlers.TransitEncrypt())
	e.POST("/transit/decrypt", handlers.TransitDecrypt())

	// start the server
	e.Logger.Fatal(e.Start(":8000"))
}
