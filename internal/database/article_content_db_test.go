package database

import (
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
}
