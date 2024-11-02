package router

import (
	"fmt"
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
	routePath := strings.TrimSuffix(filename, ".templ")

	// Convert Windows path separators to forward slashes
	routePath = filepath.ToSlash(routePath)

	var params []string
	// Handle dynamic parameters (e.g., [slug])
	segments := strings.Split(routePath, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			// Extract parameter name without brackets
			param := segment[1 : len(segment)-1]
			params = append(params, param)
			// Replace [param] with :param in route path
			segments[i] = ":" + param
		}
	}

	// Handle index routes
	if segments[len(segments)-1] == "index" {
		segments = segments[:len(segments)-1]
	}

	// Reconstruct route path
	routePath = "/" + strings.Join(segments, "/")
	if routePath == "/" {
		routePath = ""
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

	e.GET(path, func(c echo.Context) error {
		switch {
		case strings.HasSuffix(route.TemplatePath, "index.templ"):
			return pages.Index().Render(c.Request().Context(), c.Response().Writer)

		case strings.Contains(route.TemplatePath, "blog.[slug].templ"):
			slug := c.Param("slug")
			return pages.BlogPost(slug).Render(c.Request().Context(), c.Response().Writer)

		default:
			return fmt.Errorf("unknown template: %s", route.TemplatePath)
		}
	})

	return nil
}
