package content

import (
	"flag"
	"fmt"
	"log"

	"go.quinn.io/ccf/internal/codegen"
)

func Main() {
	contentDir := flag.String("content", "", "Path to content directory")
	flag.Parse()

	if *contentDir == "" {
		log.Fatal("Content directory path is required")
	}

	generator := codegen.NewContent(*contentDir)
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	fmt.Printf("Successfully generated content code at %s/fs.go\n", *contentDir)
}
