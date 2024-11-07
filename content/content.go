package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
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

// LoadItems loads all content items for a given type T from the provided filesystem.
// The items will be loaded from the specified directory.
func LoadItems[T any](fsys fs.FS, dirName string) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	delete(store, t)

	var items []ContentItem[T]

	err := fs.WalkDir(fsys, dirName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d == nil {
				return nil // Skip if directory doesn't exist
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

		// Convert markdown to HTML
		markdown := goldmark.New(
			goldmark.WithExtensions(
				&markdownImages{
					parentPath: filepath.Dir(filepath.Join("/content", path)),
				},
				highlighting.NewHighlighting(
					highlighting.WithStyle("rrt"),
				),
			),
		)
		var htmlWriter bytes.Buffer
		markdown.Convert(remainder, &htmlWriter)
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
