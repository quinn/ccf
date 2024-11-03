package content

import (
	"strings"
	"testing"
)

type Post struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
}

func TestGetItemsWithoutLoading(t *testing.T) {
	// Try to get items without loading first
	items, err := GetItems[Post]()
	if err == nil {
		t.Fatal("Expected error when getting items without loading first")
	}
	if items != nil {
		t.Fatal("Expected nil items when getting items without loading first")
	}
}

func TestLoadAndGetItems(t *testing.T) {
	// Load items first
	err := LoadItems[Post]("example/content")
	if err != nil {
		t.Fatalf("Failed to load items: %v", err)
	}

	// Get items for the Post type
	items, err := GetItems[Post]()
	if err != nil {
		t.Fatalf("Failed to get items: %v", err)
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

	expectedContent := "This is the content of Some Post.\n\n## It is markdown."
	if !strings.Contains(item.Content, expectedContent) {
		t.Errorf("Expected content to contain '%s', got '%s'", expectedContent, item.Content)
	}

	if !strings.Contains(item.HTML, "<h2>It is markdown.</h2>") {
		t.Error("Expected HTML to contain markdown conversion")
	}
}

func TestLoadItemsNonexistentDirectory(t *testing.T) {
	err := LoadItems[Post]("nonexistent")
	if err != nil {
		t.Fatal("Expected no error when loading from nonexistent directory")
	}

	items, err := GetItems[Post]()
	if err == nil {
		t.Fatal("Expected error when getting items after loading from nonexistent directory")
	}
	if items != nil {
		t.Fatal("Expected nil items when getting items after loading from nonexistent directory")
	}
}
