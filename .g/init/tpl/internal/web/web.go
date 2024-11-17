package web

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.quinn.io/ccf/assets"
	"{{ .mod }}/content"
	"{{ .mod }}/internal/router"
)

//go:embed public
var assetsFS embed.FS

func Run() {
	// Load content before starting server
	content.Initialize()

	e := echo.New()
	e.Use(middleware.Logger())

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
