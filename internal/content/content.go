package content

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

// DiscoverTypes uses reflection to discover content types from a config file
func (s *Store) DiscoverTypes(configPath string) error {
	// Parse the Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, configPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get package path for the config file
	pkgPath := filepath.Dir(configPath)

	// Create a map to store type definitions
	typeSpecs := make(map[string]*ast.TypeSpec)

	// Find all type declarations
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					typeSpecs[typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}

	// Load the package using go/types
	pkg, err := loadPackage(pkgPath)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	// Register each type
	for typeName, typeSpec := range typeSpecs {
		// Skip if not a struct
		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			continue
		}

		// Get the actual type from the package
		obj := pkg.Scope().Lookup(typeName)
		if obj == nil {
			continue
		}

		// Create a new instance of the type using reflection
		typ := reflect.TypeOf(struct{}{})
		if named, ok := obj.Type().(*types.Named); ok {
			if _, ok := named.Underlying().(*types.Struct); ok {
				// Create a new struct type with the same fields
				fields := make([]reflect.StructField, 0)
				for i := 0; i < named.Underlying().(*types.Struct).NumFields(); i++ {
					field := named.Underlying().(*types.Struct).Field(i)
					tag := named.Underlying().(*types.Struct).Tag(i)
					fields = append(fields, reflect.StructField{
						Name: field.Name(),
						Type: reflect.TypeOf(""), // Assuming all fields are strings for now
						Tag:  reflect.StructTag(tag),
					})
				}
				typ = reflect.StructOf(fields)
			}
		}

		// Register the type
		folderName := strings.ToLower(typeName) + "s"
		s.types[typeName] = &ContentType{
			Type:       typ,
			FolderName: folderName,
		}
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
			html := markdown.ToHTML(remainder, nil, nil)

			item := ContentItem{
				Meta:    reflect.ValueOf(meta).Elem().Interface(),
				Content: string(remainder),
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

// GetItems returns all items for a given content type name
func (s *Store) GetItems(typeName string) ([]ContentItem, error) {
	ct, ok := s.types[typeName]
	if !ok {
		return nil, fmt.Errorf("content type %s not registered", typeName)
	}

	return ct.Items, nil
}

// GetContentTypes returns all registered content type names
func (s *Store) GetContentTypes() []string {
	types := make([]string, 0, len(s.types))
	for typeName := range s.types {
		types = append(types, typeName)
	}
	return types
}
