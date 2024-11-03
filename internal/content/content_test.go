package content

import (
	"path/filepath"
	"testing"
)

func TestContentStore(t *testing.T) {
	store := New()

	// Discover types from the example config
	configPath := filepath.Join("..", "..", "example", "content", "config.go")
	err := store.DiscoverTypes(configPath)
	if err != nil {
		t.Fatalf("Failed to discover types: %v", err)
	}

	// Verify Post type was discovered
	types := store.GetContentTypes()
	if len(types) != 1 || types[0] != "Post" {
		t.Fatalf("Expected to find Post type, got %v", types)
	}

	// Load content from the example directory
	err = store.Load("../../example/content")
	if err != nil {
		t.Fatalf("Failed to load content: %v", err)
	}

	// Get items for the Post type
	items, err := store.GetItems("Post")
	if err != nil {
		t.Fatalf("Failed to get items: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Expected at least one item")
	}

	// Check the first item
	item := items[0]
	post, ok := item.Meta.(struct {
		Title       string `yaml:"title"`
		Date        string `yaml:"date"`
		Description string `yaml:"description"`
	})
	if !ok {
		t.Fatal("Failed to cast item meta to Post type")
	}

	if post.Title != "Some Post" {
		t.Errorf("Expected title 'Some Post', got '%s'", post.Title)
	}

	if post.Date != "2014-01-06" {
		t.Errorf("Expected date '2014-01-06', got '%s'", post.Date)
	}

	if post.Description != "Brief description of some post" {
		t.Errorf("Expected description 'Brief description of some post', got '%s'", post.Description)
	}

	if item.HTML == "" {
		t.Error("Expected non-empty HTML content")
	}
}
