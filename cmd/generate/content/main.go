package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ContentLoader handles loading and managing content types
type ContentLoader struct {
	contentDir string
	dirs       []string
}

// NewContentLoader initializes a new ContentLoader with the given content directory
func NewContentLoader(contentDir string) (*ContentLoader, error) {
	dirs, err := getContentDirs(contentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize content loader: %w", err)
	}

	return &ContentLoader{
		contentDir: contentDir,
		dirs:       dirs,
	}, nil
}

// GetContentTypes returns all available content type directories
func (cl *ContentLoader) GetContentTypes() []string {
	return cl.dirs
}

// GenerateCode generates the content filesystem code
func (cl *ContentLoader) GenerateCode(outputPath string) error {
	// Create template data
	data := struct {
		Dirs string
	}{
		Dirs: strings.Join(cl.dirs, " "),
	}

	// Read template file
	tmpl, err := template.ParseFiles("internal/codegen/templates/content.gotmpl")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// getContentDirs returns a list of content type directories
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

	loader, err := NewContentLoader("example/content")
	if err != nil {
		log.Fatalf("Failed to initialize content loader: %v", err)
	}

	if err := loader.GenerateCode(*output); err != nil {
		log.Fatalf("Failed to generate content code: %v", err)
	}

	fmt.Printf("Successfully generated content code at %s\n", *output)
}
