package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"go.quinn.io/go-astro/internal/router"
)

func main() {
	e := echo.New()

	pagesDir := filepath.Join(".", "pages")
	if err := router.GenerateRoutes(e, pagesDir); err != nil {
		log.Fatalf("failed to generate routes: %v", err)
	}

	fmt.Println("Server starting on http://localhost:3000")
	if err := e.Start(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
