package content

import (
	"fmt"
	"io/fs"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
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

// LoadItems loads all content items for a given type T from the specified content directory.
// The items will be loaded from a subdirectory matching the lowercase type name + "s"
// (e.g., Post -> posts).
func LoadItems[T any](_ string) error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	delete(store, t)

	// Determine folder name from type name (e.g., Post -> posts)
	folderName := strings.ToLower(t.Name()) + "s"

	var items []ContentItem[T]

	err := fs.WalkDir(contentFS, folderName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d == nil {
				return nil // Skip if directory doesn't exist
			}
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := fs.ReadFile(contentFS, path)
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
		html := markdown.ToHTML(remainder, nil, nil)

		// Get relative path without extension for routing
		relPath := strings.TrimSuffix(strings.TrimPrefix(path, folderName+"/"), ".md")

		// Handle index files by removing the /index suffix
		if strings.HasSuffix(relPath, "/index") {
			relPath = strings.TrimSuffix(relPath, "/index")
		}

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
