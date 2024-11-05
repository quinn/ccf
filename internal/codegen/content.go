package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ContentType struct {
	Name   string
	Fields []*ast.Field
}

type ContentGenerator struct {
	ContentDir string
}

func NewContentGenerator(contentDir string) *ContentGenerator {
	return &ContentGenerator{
		ContentDir: contentDir,
	}
}

func (g *ContentGenerator) parseContentTypes(configPath string) ([]ContentType, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, configPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	var types []ContentType
	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				types = append(types, ContentType{
					Name:   typeSpec.Name.Name,
					Fields: structType.Fields.List,
				})
			}
		}
		return true
	})

	return types, nil
}

func (g *ContentGenerator) getContentDirs() ([]string, error) {
	entries, err := os.ReadDir(g.ContentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read content directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

var templateFuncs = template.FuncMap{
	"lower": strings.ToLower,
}

func (g *ContentGenerator) Generate() error {
	// Get content types from config
	types, err := g.parseContentTypes(filepath.Join(g.ContentDir, "config.go"))
	if err != nil {
		return fmt.Errorf("failed to parse content types: %w", err)
	}

	// Get content directories
	dirs, err := g.getContentDirs()
	if err != nil {
		return fmt.Errorf("failed to get content directories: %w", err)
	}

	// Create template data
	data := struct {
		Types []ContentType
		Dirs  string
	}{
		Types: types,
		Dirs:  strings.Join(dirs, " "),
	}

	// Read template file
	tmplContent, err := templates.ReadFile("templates/content.gotmpl")
	if err != nil {
		return fmt.Errorf("failed to read content template: %w", err)
	}

	tmpl, err := template.New("content.gotmpl").Funcs(templateFuncs).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Set output path relative to content directory
	output := filepath.Join(g.ContentDir, "fs.go")

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
