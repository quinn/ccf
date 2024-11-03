package content

import (
	"testing"
)

type TestPost struct {
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}

func TestContentStore(t *testing.T) {
	store := New()

	// Register the test post type
	err := store.RegisterType(TestPost{})
	if err != nil {
		t.Fatalf("Failed to register type: %v", err)
	}

	// Load content from the example directory
	err = store.Load("../../example/content")
	if err != nil {
		t.Fatalf("Failed to load content: %v", err)
	}

	// Get items for the test post type
	items, err := store.GetItems(TestPost{})
	if err != nil {
		t.Fatalf("Failed to get items: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Expected at least one item")
	}

	// Check the first item
	item := items[0]
	post, ok := item.Meta.(TestPost)
	if !ok {
		t.Fatal("Failed to cast item meta to TestPost")
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
