package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.quinn.io/go-astro/internal/router"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	pagesDir := filepath.Join(".", "pages")
	if err := router.GenerateRoutes(e, pagesDir); err != nil {
		log.Fatalf("failed to generate routes: %v", err)
	}

	// Print registered routes
	for _, route := range e.Routes() {
		fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	}

	fmt.Println("Server starting on http://localhost:3000")
	if err := e.Start(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
