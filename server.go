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
			// MUST change to true in production to only allow https requests!
			csrf.Secure(false),
		)))

	// file routing
	e.Static("/", "public")

	// API routing - wrapper around vault API
	e.GET("/api/health", handlers.VaultHealth())

	e.GET("/api/login/csrf", handlers.FetchCSRF())
	e.POST("/api/login", handlers.Login())
	e.POST("/api/login/renew-self", handlers.RenewSelf())

	e.GET("/api/users", handlers.GetUsers())
	e.GET("/api/users/csrf", handlers.FetchCSRF())
	e.GET("/api/tokencount", handlers.GetTokenCount())
	e.GET("/api/users/role", handlers.GetRole())
	e.GET("/api/users/listroles", handlers.ListRoles())
	e.POST("/api/users/revoke", handlers.DeleteUser())
	e.POST("/api/users/create", handlers.CreateUser())

	e.GET("/api/policy", handlers.GetPolicy())
	e.DELETE("/api/policy", handlers.DeletePolicy())

	e.GET("/api/transit", handlers.TransitInfo())
	e.POST("/api/transit/encrypt", handlers.EncryptString())
	e.POST("/api/transit/decrypt", handlers.DecryptString())

	e.GET("/api/mounts", handlers.GetMounts())
	e.GET("/api/mounts/:mountname", handlers.GetMount())
	e.POST("/api/mounts/:mountname", handlers.ConfigMount())

	e.GET("/api/secrets", handlers.GetSecrets())
	e.POST("/api/secrets", handlers.PostSecrets())

	e.GET("/api/bulletins", handlers.GetBulletins())

	// start the server in HTTP
	// MUST change to HTTPS in production!
	e.Logger.Fatal(e.Start(":8000"))
}
