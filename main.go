package main

import (
	"log"
	"os"

	"go.quinn.io/ccf/cmd/esm"
	"go.quinn.io/ccf/cmd/fonts"
	"go.quinn.io/ccf/cmd/generate/content"
	"go.quinn.io/ccf/cmd/generate/pages"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("command is required")
	}

	cmd := os.Args[1]
	// Remove the command from os.Args so subcommands can parse their own flags
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch cmd {
	case "generate/content":
		content.Main()
	case "generate/pages":
		pages.Main()
	case "fonts":
		fonts.Main()
	case "esm":
		esm.Main()
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}
