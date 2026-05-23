package web

import (
	"embed"
	"fmt"
	"log"
	"os"

	"ccf/example/content"
	"ccf/example/internal/router"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.quinn.io/ccf/v2/assets"
)

//go:embed public
var assetsFS embed.FS

func Run() {
	// Load content before starting server

	e := echo.New()
	e.Use(middleware.RequestLogger())
	content.InitializePost(e)

	// Register routes from generated code
	router.RegisterRoutes(e)

	// Attach public assets
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
