package feed_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/models"
)

// Test that FetchAll cancels promptly when context is cancelled
func TestFetchAll_RespectsCancellation(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	// create one slow server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>t</title></channel></rss>`))
	}))
	defer srv.Close()

	_, err = db.AddFeed(&models.Feed{Title: "slow", URL: srv.URL})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	f := ff.NewFetcher(db, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		f.FetchAll(ctx)
		close(done)
	}()

	// cancel quickly
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(3 * time.Second):
		t.Fatalf("FetchAll did not return after cancellation")
	}
}
