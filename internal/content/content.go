package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/gomarkdown/markdown"
)

// ContentItem represents a single content item with its metadata and rendered content
type ContentItem struct {
	Meta    interface{}
	Content string
	HTML    string
	Path    string
}

var store = make(map[reflect.Type][]ContentItem)

// GetItems returns all content items for a given type T
func GetItems[T any]() ([]ContentItem, error) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	// If items haven't been loaded yet, load them
	if _, ok := store[t]; !ok {
		if err := loadItems[T](); err != nil {
			return nil, fmt.Errorf("failed to load items: %w", err)
		}
	}

	return store[t], nil
}

// loadItems loads all content items for a given type T
func loadItems[T any]() error {
	t := reflect.TypeOf((*T)(nil)).Elem()

	// Determine folder name from type name (e.g., Post -> posts)
	folderName := strings.ToLower(t.Name()) + "s"
	contentDir := filepath.Join("content", folderName)

	// Skip if directory doesn't exist
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		return nil
	}

	var items []ContentItem

	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
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
		relPath := strings.TrimSuffix(strings.TrimPrefix(path, contentDir+"/"), ".md")

		item := ContentItem{
			Meta:    reflect.ValueOf(meta).Elem().Interface(),
			Content: string(remainder),
			HTML:    string(html),
			Path:    relPath,
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
