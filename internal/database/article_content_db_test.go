package database

import (
	"fmt"
	"testing"
)

func TestArticleContentCache(t *testing.T) {
	// Create a test database
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.DB.Close()

	// Initialize schema
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Helper: create a parent article row so FK constraint is satisfied.
	// Requires a parent feed row because articles.feed_id REFERENCES feeds(id).
	_, err = db.Exec(`INSERT INTO feeds (id, title, url) VALUES (1, 'Test Feed', 'https://example.com/feed')`)
	if err != nil {
		t.Fatalf("Failed to create test feed: %v", err)
	}

	createArticle := func(id int64) {
		_, err := db.Exec(`INSERT INTO articles (id, feed_id, title, url) VALUES (?, 1, ?, ?)`,
			id, "Test Article", fmt.Sprintf("https://example.com/%d", id))
		if err != nil {
			t.Fatalf("Failed to create article %d: %v", id, err)
		}
	}

	t.Run("GetArticleContent - not found", func(t *testing.T) {
		content, found, err := db.GetArticleContent(999)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if found {
			t.Error("Expected content to not be found")
		}
		if content != "" {
			t.Error("Expected empty content")
		}
	})

	t.Run("Set and Get ArticleContent", func(t *testing.T) {
		articleID := int64(1)
		createArticle(articleID)
		testContent := "<p>This is test article content</p>"

		// Set content
		if err := db.SetArticleContent(articleID, testContent); err != nil {
			t.Errorf("Failed to set article content: %v", err)
		}

		// Get content
		content, found, err := db.GetArticleContent(articleID)
		if err != nil {
			t.Errorf("Failed to get article content: %v", err)
		}
		if !found {
			t.Error("Expected content to be found")
		}
		if content != testContent {
			t.Errorf("Content mismatch: got %q, want %q", content, testContent)
		}
	})

	t.Run("Update existing ArticleContent", func(t *testing.T) {
		articleID := int64(2)
		createArticle(articleID)
		initialContent := "<p>Initial content</p>"
		updatedContent := "<p>Updated content</p>"

		// Set initial content
		if err := db.SetArticleContent(articleID, initialContent); err != nil {
			t.Errorf("Failed to set initial content: %v", err)
		}

		// Verify initial content
		content, found, err := db.GetArticleContent(articleID)
		if err != nil || !found || content != initialContent {
			t.Error("Failed to retrieve initial content")
		}

		// Update content (should overwrite)
		if err := db.SetArticleContent(articleID, updatedContent); err != nil {
			t.Errorf("Failed to update content: %v", err)
		}

		// Verify updated content
		content, found, err = db.GetArticleContent(articleID)
		if err != nil {
			t.Errorf("Failed to get updated content: %v", err)
		}
		if !found {
			t.Error("Expected updated content to be found")
		}
		if content != updatedContent {
			t.Errorf("Updated content mismatch: got %q, want %q", content, updatedContent)
		}
	})

	t.Run("Delete ArticleContent", func(t *testing.T) {
		articleID := int64(3)
		createArticle(articleID)
		testContent := "<p>Content to delete</p>"

		// Set content
		if err := db.SetArticleContent(articleID, testContent); err != nil {
			t.Errorf("Failed to set content: %v", err)
		}

		// Verify it exists
		_, found, err := db.GetArticleContent(articleID)
		if err != nil || !found {
			t.Error("Content should exist before deletion")
		}

		// Delete content
		if err := db.DeleteArticleContent(articleID); err != nil {
			t.Errorf("Failed to delete content: %v", err)
		}

		// Verify it's deleted
		_, found, err = db.GetArticleContent(articleID)
		if err != nil {
			t.Errorf("Error after deletion: %v", err)
		}
		if found {
			t.Error("Content should not exist after deletion")
		}
	})

	t.Run("CleanupOldArticleContents", func(t *testing.T) {
		// This test verifies the cleanup function works
		// Note: In an in-memory database, we can't test actual time-based cleanup
		// but we can verify the function executes without error
		affected, err := db.CleanupOldArticleContents(30)
		if err != nil {
			t.Errorf("CleanupOldArticleContents failed: %v", err)
		}
		// In a fresh database, should affect 0 rows
		if affected != 0 {
			t.Errorf("Expected 0 rows affected, got %d", affected)
		}
	})

	t.Run("GetArticleContentsBatch", func(t *testing.T) {
		// Setup test data — must create parent article rows for FK constraint
		testData := map[int64]string{
			10: "<p>Content for article 10</p>",
			20: "<p>Content for article 20</p>",
			30: "<p>Content for article 30</p>",
		}

		for articleID := range testData {
			createArticle(articleID)
		}

		// Set contents for multiple articles
		for articleID, content := range testData {
			if err := db.SetArticleContent(articleID, content); err != nil {
				t.Fatalf("Failed to set content for article %d: %v", articleID, err)
			}
		}

		// Test 1: Get all existing contents
		articleIDs := []int64{10, 20, 30}
		contents, err := db.GetArticleContentsBatch(articleIDs)
		if err != nil {
			t.Fatalf("GetArticleContentsBatch failed: %v", err)
		}

		// Verify we got all 3 contents
		if len(contents) != 3 {
			t.Errorf("Expected 3 contents, got %d", len(contents))
		}

		// Verify each content
		for articleID, expectedContent := range testData {
			if content, ok := contents[articleID]; !ok {
				t.Errorf("Missing content for article %d", articleID)
			} else if content != expectedContent {
				t.Errorf("Content mismatch for article %d: got %q, want %q", articleID, content, expectedContent)
			}
		}

		// Test 2: Mix of existing and non-existing articles
		mixedIDs := []int64{10, 999, 20} // 999 doesn't exist
		contents, err = db.GetArticleContentsBatch(mixedIDs)
		if err != nil {
			t.Fatalf("GetArticleContentsBatch with mixed IDs failed: %v", err)
		}

		if len(contents) != 2 {
			t.Errorf("Expected 2 contents for mixed IDs, got %d", len(contents))
		}

		if _, ok := contents[999]; ok {
			t.Error("Should not have content for non-existent article 999")
		}

		// Test 3: Empty slice
		contents, err = db.GetArticleContentsBatch([]int64{})
		if err != nil {
			t.Fatalf("GetArticleContentsBatch with empty slice failed: %v", err)
		}
		if len(contents) != 0 {
			t.Errorf("Expected empty map for empty input, got %d entries", len(contents))
		}
	})
}
