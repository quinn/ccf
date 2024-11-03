package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func main() {
	contentDir := flag.String("content", "", "Directory containing content files")
	output := flag.String("output", "internal/content/generated.go", "Output path for generated content code")
	flag.Parse()

	if *contentDir == "" {
		log.Fatal("Content directory is required")
	}

	// Create template data
	data := struct {
		ContentDir string
	}{
		ContentDir: *contentDir,
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
