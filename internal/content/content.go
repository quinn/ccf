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

// ContentType represents a content type with its metadata and rendered content
type ContentType struct {
	Type       reflect.Type
	FolderName string
	Items      []ContentItem
}

// ContentItem represents a single content item with its metadata and rendered content
type ContentItem struct {
	Meta    interface{}
	Content string
	HTML    string
	Path    string
}

// Store manages content types and their items
type Store struct {
	types map[string]*ContentType
}

// New creates a new content store
func New() *Store {
	return &Store{
		types: make(map[string]*ContentType),
	}
}

// RegisterType registers a new content type based on a struct
func (s *Store) RegisterType(example interface{}) error {
	t := reflect.TypeOf(example)
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("content type must be a struct")
	}

	// Get the type name and convert to lowercase for folder name
	typeName := t.Name()
	folderName := strings.ToLower(typeName) + "s"

	s.types[typeName] = &ContentType{
		Type:       t,
		FolderName: folderName,
	}

	return nil
}

// Load loads all content for registered types from a base directory
func (s *Store) Load(baseDir string) error {
	for _, ct := range s.types {
		contentDir := filepath.Join(baseDir, ct.FolderName)

		// Skip if directory doesn't exist
		if _, err := os.Stat(contentDir); os.IsNotExist(err) {
			continue
		}

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
			meta := reflect.New(ct.Type).Interface()

			// Parse frontmatter
			remainder, err := frontmatter.Parse(strings.NewReader(string(content)), meta)
			if err != nil {
				return fmt.Errorf("failed to parse frontmatter in %s: %w", path, err)
			}

			// Convert markdown to HTML
			html := markdown.ToHTML([]byte(remainder), nil, nil)

			item := ContentItem{
				Meta:    reflect.ValueOf(meta).Elem().Interface(),
				Content: remainder,
				HTML:    string(html),
				Path:    path,
			}

			ct.Items = append(ct.Items, item)
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to load content type %s: %w", ct.Type.Name(), err)
		}
	}

	return nil
}

// GetItems returns all items for a given content type
func (s *Store) GetItems(contentType interface{}) ([]ContentItem, error) {
	t := reflect.TypeOf(contentType)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("content type must be a struct")
	}

	ct, ok := s.types[t.Name()]
	if !ok {
		return nil, fmt.Errorf("content type %s not registered", t.Name())
	}

	return ct.Items, nil
}
