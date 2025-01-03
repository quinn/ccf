package codegen

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*.gotmpl
var templates embed.FS

// PageRoute represents a route generated from a page file
type PageRoute struct {
	Path         string
	TemplatePath string
	GETHandler   string
	POSTHandler  string
	HasPOST      bool
	Params       []string
	Component    string
}

// Generator handles the code generation for routes
type Generator struct {
	PagesDir    string
	OutputPath  string
	PackageName string
}

// New creates a new Generator instance
func New(pagesDir, outputPath, packageName string) *Generator {
	return &Generator{
		PagesDir:    pagesDir,
		OutputPath:  outputPath,
		PackageName: packageName,
	}
}

// Generate scans the pages directory and generates route code
func (g *Generator) Generate() error {
	routes, err := g.scanPagesDirectory()
	if err != nil {
		return fmt.Errorf("failed to scan pages directory: %w", err)
	}

	return g.generateRouterCode(routes)
}

// scanPagesDirectory walks through the pages directory and generates route information
func (g *Generator) scanPagesDirectory() ([]PageRoute, error) {
	var routes []PageRoute

	err := filepath.Walk(g.PagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".templ") {
			return nil
		}

		// Get relative path from pages directory
		relPath, err := filepath.Rel(g.PagesDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		route, err := g.parseRouteFromFilename(relPath)
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

// toUpperCamelCase converts a string to upper camel case
func toUpperCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	var result strings.Builder
	for _, word := range words {
		if word == "" {
			continue
		}
		result.WriteString(strings.ToUpper(word[:1]) + strings.ToLower(word[1:]))
	}
	return result.String()
}

// parseRouteFromFilename converts a template filename into a route
func (g *Generator) parseRouteFromFilename(filename string) (PageRoute, error) {
	// Remove .templ suffix and get the base filename without directory
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, ".templ")

	var params []string
	var routeParts []string
	var handlerParts []string

	// Split the path into segments by periods
	segments := strings.Split(base, ".")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			// Extract parameter name without brackets
			param := segment[1 : len(segment)-1]
			params = append(params, param)
			routeParts = append(routeParts, ":"+param)
			// Convert parameter to upper camel case for handler name
			handlerParts = append(handlerParts, toUpperCamelCase(param))
		} else {
			// Special case for index
			if segment == "index" && i == 0 {
				routeParts = append(routeParts, "")
			} else {
				routeParts = append(routeParts, segment)
			}
			handlerParts = append(handlerParts, toUpperCamelCase(segment))
		}
	}

	routePath := "/" + strings.Join(routeParts, "/")

	// Clean up the route path
	routePath = strings.ReplaceAll(routePath, "//", "/")
	if routePath != "/" && strings.HasSuffix(routePath, "/") {
		routePath = strings.TrimSuffix(routePath, "/")
	}

	// Generate handler name by combining all parts and adding Handler suffix
	component := strings.Join(handlerParts, "")
	getHandler := component + "GET"
	postHandler := component + "POST"

	// Check if the POST handler exists in the template file
	hasPost := g.hasPostHandler(filename, component)

	return PageRoute{
		Path:         routePath,
		TemplatePath: filename,
		GETHandler:   getHandler,
		POSTHandler:  postHandler,
		HasPOST:      hasPost,
		Params:       params,
		Component:    component,
	}, nil
}

// hasPostHandler checks if a templ file has a POST handler function
func (g *Generator) hasPostHandler(filename string, component string) bool {
	fullPath := filepath.Join(g.PagesDir, filename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return false
	}

	// Look for a function named [component]POST
	postHandlerPattern := fmt.Sprintf("func %sPOST", component)
	return strings.Contains(string(content), postHandlerPattern)
}

// generateRouterCode generates the router implementation
func (g *Generator) generateRouterCode(routes []PageRoute) error {
	tmplContent, err := templates.ReadFile("templates/router.gotmpl")
	if err != nil {
		return fmt.Errorf("failed to read router template: %w", err)
	}

	tmpl, err := template.New("router").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(g.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(g.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	data := struct {
		PackageName string
		Routes      []PageRoute
	}{
		PackageName: g.PackageName,
		Routes:      routes,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
