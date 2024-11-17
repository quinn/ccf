package main

import (
	"flag"
	"log"
	"path/filepath"

	"go.quinn.io/ccf/internal/codegen"
)

func main() {
	pagesDir := flag.String("pages", "pages", "Directory containing page templates")
	output := flag.String("output", "internal/router/generated.go", "Output path for generated router code")
	pkgName := flag.String("package", "router", "Package name for generated code")
	flag.Parse()

	absPages, err := filepath.Abs(*pagesDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for pages directory: %v", err)
	}

	generator := codegen.New(absPages, *output, *pkgName)
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate router code: %v", err)
	}

	log.Printf("Successfully generated router code at %s", *output)
}
