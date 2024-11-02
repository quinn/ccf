package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.quinn.io/go-astro/example/router"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	// Register routes from generated code
	router.RegisterRoutes(e)

	// Print registered routes
	for _, route := range e.Routes() {
		fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	}

	fmt.Println("Server starting on http://localhost:3000")
	if err := e.Start(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
