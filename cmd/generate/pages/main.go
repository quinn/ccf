package pages

import (
	"flag"
	"log"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/mod/modfile"

	"go.quinn.io/ccf/internal/codegen"
)

func Main() {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		panic(err)
	}

	pagesDir := flag.String("pages", "pages", "Directory containing page templates")
	output := flag.String("output", "internal/router/generated.go", "Output path for generated router code")
	pkgName := flag.String("package", "router", "Package name for generated code")
	flag.Parse()

	absPages, err := filepath.Abs(*pagesDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for pages directory: %v", err)
	}

	pagesImport := path.Join(mod.Module.Mod.Path, *pagesDir)

	generator := codegen.NewPages(absPages, *output, *pkgName, pagesImport)
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate router code: %v", err)
	}

	log.Printf("Successfully generated router code at %s", *output)
}
