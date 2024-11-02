package router

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"go.quinn.io/go-astro/pages"
)

// PageRoute represents a route generated from a page file
type PageRoute struct {
	Path         string
	TemplatePath string
	Params       []string
}

// GenerateRoutes scans the pages directory and generates Echo routes
func GenerateRoutes(e *echo.Echo, pagesDir string) error {
	routes, err := scanPagesDirectory(pagesDir)
	if err != nil {
		return fmt.Errorf("failed to scan pages directory: %w", err)
	}

	for _, route := range routes {
		if err := registerRoute(e, route); err != nil {
			return fmt.Errorf("failed to register route %s: %w", route.Path, err)
		}
		log.Printf("Registered route: %s (template: %s)", route.Path, route.TemplatePath)
	}

	return nil
}

// scanPagesDirectory walks through the pages directory and generates route information
func scanPagesDirectory(pagesDir string) ([]PageRoute, error) {
	var routes []PageRoute

	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".templ") {
			return nil
		}

		// Get relative path from pages directory
		relPath, err := filepath.Rel(pagesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		route, err := parseRouteFromFilename(relPath)
		if err != nil {
			return fmt.Errorf("failed to parse route from %s: %w", relPath, err)
		}

		routes = append(routes, route)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan pages directory: %w", err)
	}

	return routes, nil
}

// parseRouteFromFilename converts a template filename into a route
func parseRouteFromFilename(filename string) (PageRoute, error) {
	// Remove .templ suffix
	templatePath := strings.TrimSuffix(filename, ".templ")

	var params []string
	var routePath string
	var argName string
	var brackets bool

	// Parse the path and extract parameters
	for _, char := range templatePath {
		switch char {
		case '[':
			brackets = true
		case ']':
			brackets = false
			params = append(params, argName)
			routePath += ":" + argName
			argName = ""
		default:
			if brackets {
				argName += string(char)
			} else {
				routePath += string(char)
			}
		}
	}

	// Handle index routes
	if strings.HasSuffix(routePath, "index") {
		routePath = strings.TrimSuffix(routePath, "index")
	}

	// Ensure path starts with /
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}

	// Clean up any double slashes
	routePath = strings.ReplaceAll(routePath, "//", "/")

	// Trim trailing slash unless it's the root path
	if routePath != "/" && strings.HasSuffix(routePath, "/") {
		routePath = strings.TrimSuffix(routePath, "/")
	}

	return PageRoute{
		Path:         routePath,
		TemplatePath: filename,
		Params:       params,
	}, nil
}

// registerRoute adds a route to the Echo instance
func registerRoute(e *echo.Echo, route PageRoute) error {
	path := route.Path
	if path == "" {
		path = "/"
	}

	handler := func(c echo.Context) error {
		log.Printf("Handling request for path: %s", c.Request().URL.Path)

		switch {
		case strings.HasSuffix(route.TemplatePath, "index.templ"):
			return pages.Index().Render(c.Request().Context(), c.Response().Writer)

		case strings.Contains(route.TemplatePath, "blog.[slug].templ"):
			slug := c.Param("slug")
			log.Printf("Blog post request with slug: %s", slug)
			return pages.BlogPost(slug).Render(c.Request().Context(), c.Response().Writer)

		default:
			return fmt.Errorf("unknown template: %s", route.TemplatePath)
		}
	}

	e.GET(path, handler)
	log.Printf("Registered handler for path: %s", path)
	return nil
}
