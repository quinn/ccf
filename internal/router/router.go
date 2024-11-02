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
		registerRoute(e, route)
	}

	return nil
}

// scanPagesDirectory walks through the pages directory and generates route information
func scanPagesDirectory(pagesDir string) ([]PageRoute, error) {
	var routes []PageRoute

		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".templ") {
			return nil
		}

		route, err := parseRouteFromFilename(path, pagesDir)
		if err != nil {
			return fmt.Errorf("failed to parse route from %s: %w", path, err)
		}

		routes = append(routes, route)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return routes, nil
}

// parseRouteFromFilename converts a template filename into a route
func parseRouteFromFilename(filename, pagesDir string) (PageRoute, error) {
	// Remove pages directory prefix and .templ suffix
	routePath := strings.TrimPrefix(filename, pagesDir)
	routePath = strings.TrimSuffix(routePath, ".templ")

	// Convert Windows path separators to forward slashes if needed
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

	// Reconstruct route path
	routePath = strings.Join(segments, "/")
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}

	return PageRoute{
		Path:         routePath,
		TemplatePath: filename,
		Params:       params,
	}, nil
}

// registerRoute adds a route to the Echo instance
func registerRoute(e *echo.Echo, route PageRoute) {
	e.GET(route.Path, func(c echo.Context) error {
		// TODO: Implement template rendering
		return c.String(200, fmt.Sprintf("Template: %s, Path: %s", route.TemplatePath, route.Path))
	})
}
