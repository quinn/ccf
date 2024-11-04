package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ContentType struct {
	Name   string
	Fields []*ast.Field
}

func parseContentTypes(configPath string) ([]ContentType, error) {
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

func getContentDirs(contentDir string) ([]string, error) {
	entries, err := os.ReadDir(contentDir)
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

func main() {
	output := flag.String("output", "example/content/fs.go", "Output path for generated content code")
	flag.Parse()

	// Get content types from config
	types, err := parseContentTypes("example/content/config.go")
	if err != nil {
		log.Fatalf("Failed to parse content types: %v", err)
	}

	// Get content directories
	dirs, err := getContentDirs("example/content")
	if err != nil {
		log.Fatalf("Failed to get content directories: %v", err)
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
	tmpl, err := template.ParseFiles("internal/codegen/templates/content.gotmpl")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create output file
	f, err := os.Create(*output)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Printf("Successfully generated content code at %s\n", *output)
}
