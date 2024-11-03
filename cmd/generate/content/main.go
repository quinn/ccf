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
	output := flag.String("output", "example/content/generated.go", "Output path for generated content code")
	flag.Parse()

	// Get content directories
	dirs, err := getContentDirs("example/content")
	if err != nil {
		log.Fatalf("Failed to get content directories: %v", err)
	}

	// Create template data
	data := struct {
		Dirs string
	}{
		Dirs: strings.Join(dirs, " "),
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
