package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/alecthomas/chroma/v2/formatters/html"
	obsidian "github.com/powerman/goldmark-obsidian"
	"github.com/yuin/goldmark"

	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

// ContentItem represents a single content item with its metadata and rendered content
type ContentItem[T any] struct {
	Meta    T
	Content string
	HTML    string
	Slug    string
}

var store = make(map[reflect.Type]any)

// GetItems returns all content items for a given type T.
// LoadItems must be called first to populate the store.
func GetItems[T any]() ([]ContentItem[T], error) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	items, ok := store[t].([]ContentItem[T])
	if !ok {
		return nil, fmt.Errorf("no items found for type %v, ensure LoadItems was called", t)
	}

	return items, nil
}

type loadConfig struct {
	imageCallback func(imageTag string) string
}

type LoadOpt func(*loadConfig)

func ImagePostProcess(imageCallback func(imageTag string) string) LoadOpt {
	return func(config *loadConfig) {
		config.imageCallback = imageCallback
	}
}

// LoadItems loads all content items for a given type T from the provided filesystem.
// The items will be loaded from the specified directory.
func LoadItems[T any](fsys fs.FS, dirName string, opts ...LoadOpt) error {
	cfg := loadConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	t := reflect.TypeOf((*T)(nil)).Elem()
	delete(store, t)

	var items []ContentItem[T]

	slog.Info("Loading content", "type", t, "dir", dirName)
	err := fs.WalkDir(fsys, dirName, func(path string, d fs.DirEntry, err error) error {
		slog.Debug("Walking directory", "path", path)
		if err != nil {
			if d == nil {
				return fmt.Errorf("dir is missing: %w", err)
			}
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read content file %s: %w", path, err)
		}

		// Create a new instance of the content type
		meta := reflect.New(t).Interface()

		// Parse frontmatter
		remainder, err := frontmatter.Parse(strings.NewReader(string(content)), meta)
		if err != nil {
			return fmt.Errorf("failed to parse frontmatter in %s: %w", path, err)
		}

		var htmlWriter bytes.Buffer
		var cssWriter bytes.Buffer

		// Convert markdown to HTML
		markdown := goldmark.New(
			goldmark.WithExtensions(
				obsidian.NewObsidian(),
				&markdownImages{
					parentPath: filepath.Dir(filepath.Join("/content", path)),
					callback:   cfg.imageCallback,
				},
				highlighting.NewHighlighting(
					highlighting.WithStyle("rrt"),
					highlighting.WithFormatOptions(html.WithClasses(true), html.WithAllClasses(true)),
					highlighting.WithCSSWriter(&cssWriter),
					highlighting.WithGuessLanguage(true),
				),
			),
		)

		markdown.Convert(remainder, &htmlWriter)

		htmlWriter.Write([]byte("<style>"))
		b, err := cssWriter.WriteTo(&htmlWriter)
		if err != nil {
			return fmt.Errorf("failed to write CSS to HTML: %w", err)
		}

		if b == 0 {
			slog.Warn("no CSS written")
		}
		htmlWriter.Write([]byte("</style>"))

		html := htmlWriter.Bytes()

		// Get relative path without extension for routing
		relPath := strings.TrimSuffix(strings.TrimPrefix(path, dirName+"/"), ".md")

		// Handle index files by removing the /index suffix
		relPath = strings.TrimSuffix(relPath, "/index")

		item := ContentItem[T]{
			Meta:    reflect.ValueOf(meta).Elem().Interface().(T),
			Content: string(remainder),
			HTML:    string(html),
			Slug:    relPath,
		}

		items = append(items, item)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load content items: %w", err)
	}

	store[t] = items
	return nil
}
