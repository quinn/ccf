package web

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.quinn.io/go-astro/example/content"
	"go.quinn.io/go-astro/example/router"
	"go.quinn.io/go-astro/internal/assets"
	collections "go.quinn.io/go-astro/internal/content"
)

//go:embed public
var assetsFS embed.FS

func Run() {
	// Load content before starting server
	if err := collections.LoadItems[content.Post](content.ContentDir); err != nil {
		log.Fatalf("failed to load content: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	// Register routes from generated code
	router.RegisterRoutes(e)
	assets.Attach(
		e,
		"public",
		"internal/web/public",
		assetsFS,
		os.Getenv("USE_EMBEDDED_ASSETS") == "true",
	)

	fmt.Println("Server starting on http://localhost:3000")
	if err := e.Start(":3000"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
