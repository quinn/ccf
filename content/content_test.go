package content

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

type Post struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Date        string `yaml:"date"`
}

func setupTestFS() fs.FS {
	return fstest.MapFS{
		"posts/2014/some-post.md": &fstest.MapFile{
			Data: []byte(`---
title: Some Post
date: 2014-01-06
description: Brief description of some post
---
This is the content of Some Post.

## It is markdown.`),
		},
		"posts/2024/test-1-two/index.md": &fstest.MapFile{
			Data: []byte(`---
title: Index Post
date: 2024-01-01
description: Test index post
---
This is an index post.`),
		},
		"posts/2020/images-test.md": &fstest.MapFile{
			Data: []byte(`---
title: Images Test
date: 2024-01-01
description: Test post with various image types
---
# Testing Images

1. Hotlink: ![hotlink](https://example.com/image.jpg)
2. Data URL: ![data url](data:image/png;base64,abc123)
3. Absolute path: ![absolute](/images/test.png)
4. Relative path: ![relative](./images/test.jpg)`),
		},
	}
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
	fsys := setupTestFS()

	// Load items first
	err := LoadItems[Post](fsys, "posts")
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

func TestIndex(t *testing.T) {
	fsys := setupTestFS()

	// Load items first
	err := LoadItems[Post](fsys, "posts")
	if err != nil {
		t.Fatalf("Failed to load items: %v", err)
	}

	// Get items for the Post type
	items, err := GetItems[Post]()
	if err != nil {
		t.Fatalf("Failed to get items: %v", err)
	}

	var post ContentItem[Post]

	// Find the post with the title "Index Post"
	for _, item := range items {
		if item.Meta.Title == "Index Post" {
			post = item
			break
		}
	}

	if post.Meta.Title != "Index Post" {
		t.Errorf("Expected title 'Index Post', got '%s'", post.Meta.Title)
	}

	if post.Slug != "2024/test-1-two" {
		t.Errorf("Expected slug '2024/test-1-two', got '%s'", post.Slug)
	}
}

// func TestLoadItemsNonexistentDirectory(t *testing.T) {
// 	fsys := fstest.MapFS{}

// 	err := LoadItems[Post](fsys, "posts")
// 	if err != nil {
// 		t.Fatal("Expected no error when loading from empty filesystem")
// 	}

// 	items, err := GetItems[Post]()
// 	if err == nil {
// 		t.Fatal("Expected error when getting items after loading from empty filesystem")
// 	}
// 	if items != nil {
// 		t.Fatal("Expected nil items when getting items after loading from empty filesystem")
// 	}
// }

func TestImageURLs(t *testing.T) {
	fsys := setupTestFS()

	// Load items first
	err := LoadItems[Post](fsys, "posts")
	if err != nil {
		t.Fatalf("Failed to load items: %v", err)
	}

	// Get items for the Post type
	items, err := GetItems[Post]()
	if err != nil {
		t.Fatalf("Failed to get items: %v", err)
	}

	var imagePost ContentItem[Post]
	for _, item := range items {
		if item.Meta.Title == "Images Test" {
			imagePost = item
			break
		}
	}

	if imagePost.Meta.Title == "" {
		t.Fatal("Could not find Images Test post")
	}

	// Test hotlink - should remain unchanged
	if !strings.Contains(imagePost.HTML, `<img src="https://example.com/image.jpg"`) {
		t.Error("Hotlinked image URL was modified")
	}

	// Test data URL - should remain unchanged
	if !strings.Contains(imagePost.HTML, `<img src="data:image/png;base64,abc123"`) {
		t.Error("Data URL was modified")
	}

	// Test absolute path - should remain unchanged
	if !strings.Contains(imagePost.HTML, `<img src="/images/test.png"`) {
		t.Error("Absolute path was modified")
	}

	// Test relative path - should be prefixed with parent path
	expectedPath := "/content/posts/2020/images/test.jpg"
	if !strings.Contains(imagePost.HTML, `<img src="`+expectedPath+`"`) {
		t.Errorf("Relative path not properly prefixed. Expected %s in HTML: %s", expectedPath, imagePost.HTML)
	}
}
