package content

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"go.quinn.io/ccf/internal/codegen"
)

func Main() {
	contentDir := flag.String("content", "", "Path to content directory")
	debugPtr := flag.Bool("debug", os.Getenv("DEBUG") == "true", "Enable debug logging")

	flag.Parse()

	if *contentDir == "" {
		log.Fatal("Content directory path is required")
	}

	if *debugPtr {
		h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(h))
	}

	generator := codegen.NewContent(*contentDir)
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	fmt.Printf("Successfully generated content code at %s/fs.go\n", *contentDir)
}
