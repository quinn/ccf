package assets

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// Previously I've managed this in middleware and context.
// Trying global variables for now.
var prefix string
var ffs *fingerprintedFS

func Attach(e *echo.Echo, inputprefix string, assetDir string, embedFS embed.FS, embedded bool) {
	if ffs != nil {
		log.Fatalf("assets.Attach called more than once")
	}

	var inputFS fs.FS

	if embedded {
		inputFS = echo.MustSubFS(embedFS, prefix)
	} else {
		inputFS = os.DirFS(assetDir)
	}

	fingerprintFS, err := newFFS(inputFS)
	if err != nil {
		log.Fatalf("failed to create fingerprinted FS: %s", err.Error())
	}

	// init global variables
	prefix = inputprefix
	ffs = fingerprintFS

	e.StaticFS("/"+prefix, fingerprintFS)

	e.GET("/"+prefix+"/asset-manifest.json", func(c echo.Context) error {
		manifest, err := fingerprintFS.Manifest()
		if err != nil {
			return fmt.Errorf("failed to get asset manifest: %w", err)
		}

		return c.JSON(http.StatusOK, manifest)
	})
}

func Manifest() map[string]string {
	if ffs == nil {
		panic("Please run assets.Attach before calling assets.Manifest")
	}

	m, err := ffs.Manifest()
	if err != nil {
		return nil
	}

	return m
}

func ImportMap() *template.Template {
	if ffs == nil {
		panic("Please run assets.Attach before calling assets.ImportMap")
	}

	manifest := Manifest()
	if manifest == nil {
		return nil
	}

	importMap := "<script type=\"importmap\">\n{\n"
	importMap += "\"imports\": {\n"
	var imports []string
	for path := range manifest {
		imports = append(imports, fmt.Sprintf(`  "%s": "%s"`, path, Path(path)))
	}
	importMap += fmt.Sprintf("%s\n", strings.Join(imports, ",\n"))
	importMap += "}\n"
	importMap += "}\n</script>\n"

	return template.Must(template.New("importmap").Parse(importMap))
}

func Path(path string) string {
	if ffs == nil {
		panic("Please run assets.Attach before calling assets.Path")
	}

	url, err := ffs.URL(path)
	if err != nil {
		return fmt.Sprintf("failed to get fingerprinted URL: %v", err)
	}

	return "/" + prefix + url
}

func Inline(path string) []byte {
	if ffs == nil {
		panic("Please run assets.Attach before calling assets.Inline")
	}

	data, err := ffs.ReadFile(path)
	if err != nil {
		return []byte(fmt.Sprintf("failed to read file: %v", err))
	}

	return data
}
