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

	e.GET("/api/health", handlers.VaultHealth())

	e.GET("/api/login/csrf", handlers.FetchCSRF())
	e.POST("/api/login", handlers.Login())

	e.GET("/api/users", handlers.GetUsers())
	e.POST("/api/users/revoke", handlers.DeleteUser())

	e.GET("/api/policies", handlers.ListPolicies())
	e.GET("/api/policies/:policyname", handlers.GetPolicy())
	e.DELETE("/api/policies/:policyname", handlers.DeletePolicy())

	e.GET("/api/transit", handlers.FetchCSRF())
	e.POST("/api/transit/encrypt", handlers.EncryptString())
	e.POST("/api/transit/decrypt", handlers.DecryptString())

	e.GET("/api/mounts", handlers.GetMounts())
	e.GET("/api/mounts/:mountname", handlers.GetMount())
	e.POST("/api/mounts/:mountname", handlers.ConfigMount())

	e.GET("/api/secrets", handlers.GetSecrets())

	// start the server
	e.Logger.Fatal(e.Start(":8000"))
}
