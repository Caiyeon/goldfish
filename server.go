package main

import (
	"github.com/caiyeon/goldfish/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// create a new instance of Echo
	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// routing
	e.Static("/", "public")

	e.POST("/login", handlers.Login())
	e.GET("/users", handlers.Users())
	e.DELETE("/users", handlers.DeleteUser())

	// Start a web server
	e.Logger.Fatal(e.Start(":8000"))
}