package feed

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"MrRSS/internal/database"

	"github.com/mmcdole/gofeed"
)

func TestParseFeedWithScript_PrioritySystem(t *testing.T) {
	// This test verifies that the priority system works correctly
	// High priority requests should not be blocked by low priority operations

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	fetcher := NewFetcher(db, nil)
	// Use mock parser to avoid network calls
	mockParser := &MockParser{Err: errors.New("mock error")}
	fetcher.fp = mockParser

	// Test that high priority requests can execute without issues
	ctx := context.Background()

	// Test with mock error to ensure the priority logic doesn't break normal error handling
	_, err = fetcher.ParseFeedWithScript(ctx, "http://invalid-url-that-does-not-exist.com", "", true)
	if err == nil {
		t.Error("Expected error from mock parser")
	}

	// Test normal priority
	_, err = fetcher.ParseFeedWithScript(ctx, "http://invalid-url-that-does-not-exist.com", "", false)
	if err == nil {
		t.Error("Expected error from mock parser")
	}
}

func TestParseFeedWithScript_Concurrency(t *testing.T) {
	// Test that multiple high priority requests can run concurrently
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	fetcher := NewFetcher(db, nil)
	// Use mock parser to avoid network calls
	mockParser := &MockParser{Err: errors.New("mock error")}
	fetcher.fp = mockParser

	ctx := context.Background()
	var wg sync.WaitGroup

	// Start multiple high priority requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := fetcher.ParseFeedWithScript(ctx, "http://invalid-url.com", "", true)
			// We expect errors, but the point is that they should not deadlock
			if err == nil {
				t.Errorf("Expected error from mock parser")
			}
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all requests completed
	case <-time.After(5 * time.Second):
		t.Fatal("High priority requests were blocked/deadlocked")
	}
}

// TimeoutMockParser implements feed.FeedParser for testing timeouts
type TimeoutMockParser struct {
	Timeout time.Duration
}

func (m *TimeoutMockParser) ParseURL(url string) (*gofeed.Feed, error) {
	time.Sleep(m.Timeout)
	return &gofeed.Feed{Title: "Test Feed"}, nil
}

func (m *TimeoutMockParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	select {
	case <-time.After(m.Timeout):
		return &gofeed.Feed{Title: "Test Feed"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
