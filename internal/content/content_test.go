package content

import (
	"path/filepath"
	"testing"
)

type Post struct {
	Title       string `yaml:"title"`
	Date        string `yaml:"date"`
	Description string `yaml:"description"`
}

func TestGetItems(t *testing.T) {
	// Get items for the Post type
	items, err := GetItems[Post]()
	if err != nil {
		t.Fatalf("Failed to get items: %w", err)
	}

	if len(items) == 0 {
		t.Fatal("Expected at least one item")
	}

	// Check the first item
	item := items[0]
	post := item.Meta

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
