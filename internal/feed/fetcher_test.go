package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/translation"
	"context"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

type MockParser struct {
	Feed *gofeed.Feed
	Err  error
}

func (m *MockParser) ParseURL(url string) (*gofeed.Feed, error) {
	return m.Feed, m.Err
}

func (m *MockParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	return m.Feed, m.Err
}

func TestAddSubscription(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items:       []*gofeed.Item{},
	}

	fetcher := NewFetcher(db, translation.NewMockTranslator())
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, err := db.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds failed: %v", err)
	}

	if len(feeds) != 1 {
		t.Errorf("Expected 1 feed, got %d", len(feeds))
	}
	if feeds[0].Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", feeds[0].Title)
	}
}

func TestFetchFeed(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser for AddSubscription
	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items: []*gofeed.Item{
			{
				Title:       "Test Article",
				Link:        "http://test.com/article",
				Description: "Article Description",
				Content:     "Article Content",
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	if articles[0].Title != "Test Article" {
		t.Errorf("Expected article title 'Test Article', got '%s'", articles[0].Title)
	}
}

func TestFetchFeedWithMissingTitle(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser with an item that has no title but has content
	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items: []*gofeed.Item{
			{
				Title:       "", // Missing title
				Link:        "http://test.com/article",
				Description: "",
				Content:     "This is a short content.",
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	expectedTitle := "This is a short content."
	if articles[0].Title != expectedTitle {
		t.Errorf("Expected article title '%s', got '%s'", expectedTitle, articles[0].Title)
	}
}

func TestFetchFeedWithMissingTitleLongContent(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser with an item that has no title but has long content
	longContent := "This is a very long article content that should be truncated to generate a title from the beginning of the content when the title is missing from the RSS feed item."
	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items: []*gofeed.Item{
			{
				Title:       "", // Missing title
				Link:        "http://test.com/article",
				Description: "",
				Content:     longContent,
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	// Should be truncated to 100 chars + "..."
	expectedTitle := longContent[:100] + "..."
	if articles[0].Title != expectedTitle {
		t.Errorf("Expected article title '%s', got '%s'", expectedTitle, articles[0].Title)
	}
}

func TestFetchFeedWithAudioEnclosure(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser with an item that has audio enclosure (podcast)
	mockFeed := &gofeed.Feed{
		Title:       "Test Podcast",
		Description: "Test Podcast Description",
		Items: []*gofeed.Item{
			{
				Title:       "Podcast Episode 1",
				Link:        "http://test.com/episode1",
				Description: "Episode Description",
				Content:     "Episode Content",
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/audio/episode1.mp3",
						Type:   "audio/mpeg",
						Length: "12345678",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	if articles[0].Title != "Podcast Episode 1" {
		t.Errorf("Expected article title 'Podcast Episode 1', got '%s'", articles[0].Title)
	}
	expectedAudioURL := "https://test.com/audio/episode1.mp3"
	if articles[0].AudioURL != expectedAudioURL {
		t.Errorf("Expected audio URL '%s', got '%s'", expectedAudioURL, articles[0].AudioURL)
	}
}

func TestFetchFeedWithImageEnclosure(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser with an item that has image enclosures (png and jpeg)
	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items: []*gofeed.Item{
			{
				Title:           "Article with PNG",
				Link:            "http://test.com/article1",
				Description:     "Article Description",
				Content:         "Article Content",
				PublishedParsed: &[]time.Time{time.Now().Add(time.Hour)}[0], // More recent
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/image1.png",
						Type:   "image/png",
						Length: "12345",
					},
				},
			},
			{
				Title:           "Article with JPEG",
				Link:            "http://test.com/article2",
				Description:     "Article Description",
				Content:         "Article Content",
				PublishedParsed: &[]time.Time{time.Now()}[0], // Less recent
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/image2.jpg",
						Type:   "image/jpeg",
						Length: "23456",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}

	// Check PNG image extraction
	if articles[0].ImageURL != "https://test.com/images/image1.png" {
		t.Errorf("Expected PNG image URL, got '%s'", articles[0].ImageURL)
	}

	// Check JPEG image extraction
	if articles[1].ImageURL != "https://test.com/images/image2.jpg" {
		t.Errorf("Expected JPEG image URL, got '%s'", articles[1].ImageURL)
	}
}

func TestFetchFeedWithMultipleEnclosures(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Add a feed first
	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Mock the parser with an item that has both image and audio enclosures
	mockFeed := &gofeed.Feed{
		Title:       "Test Podcast",
		Description: "Test Podcast Description",
		Items: []*gofeed.Item{
			{
				Title:       "Podcast Episode with Cover",
				Link:        "http://test.com/episode1",
				Description: "Episode Description",
				Content:     "Episode Content",
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/cover.jpg",
						Type:   "image/jpeg",
						Length: "12345",
					},
					{
						URL:    "https://test.com/audio/episode1.mp3",
						Type:   "audio/mpeg",
						Length: "98765432",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	// Fetch the feed
	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}

	// Should have both image and audio
	expectedImageURL := "https://test.com/images/cover.jpg"
	if articles[0].ImageURL != expectedImageURL {
		t.Errorf("Expected image URL '%s', got '%s'", expectedImageURL, articles[0].ImageURL)
	}

	expectedAudioURL := "https://test.com/audio/episode1.mp3"
	if articles[0].AudioURL != expectedAudioURL {
		t.Errorf("Expected audio URL '%s', got '%s'", expectedAudioURL, articles[0].AudioURL)
	}
}

func TestFetchSingleFeedProgressTracking(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db, translation.NewMockTranslator())

	// Add multiple feeds
	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items:       []*gofeed.Item{},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	feedID1, _ := fetcher.AddSubscription("http://test1.com/rss", "Category 1", "")
	feedID2, _ := fetcher.AddSubscription("http://test2.com/rss", "Category 2", "")
	feedID3, _ := fetcher.AddSubscription("http://test3.com/rss", "Category 3", "")

	feed1, _ := db.GetFeedByID(feedID1)
	feed2, _ := db.GetFeedByID(feedID2)
	feed3, _ := db.GetFeedByID(feedID3)

	// Queue multiple feeds for refresh
	go fetcher.FetchSingleFeed(context.Background(), *feed1)
	go fetcher.FetchSingleFeed(context.Background(), *feed2)
	go fetcher.FetchSingleFeed(context.Background(), *feed3)

	// Give goroutines time to start and queue
	time.Sleep(50 * time.Millisecond)

	// Check progress during execution
	// Note: Due to fast execution with mock parser, progress might already be partially complete
	progress := fetcher.GetProgress()
	// We should see that at least some feeds were queued (total should be 1-3)
	if progress.Total < 1 || progress.Total > 3 {
		t.Errorf("Expected total between 1-3, got %d", progress.Total)
	}

	// Wait for all feeds to complete
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for feeds to complete")
		case <-ticker.C:
			progress := fetcher.GetProgress()
			if !progress.IsRunning {
				// Check that we processed all 3 feeds
				if progress.Current != 3 {
					t.Errorf("Expected current to be 3, got %d", progress.Current)
				}
				return
			}
		}
	}
}

func TestFetchSingleFeedDuplicatePrevention(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db, translation.NewMockTranslator())

	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items:       []*gofeed.Item{},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	feedID, _ := fetcher.AddSubscription("http://test.com/rss", "Category", "")
	feed, _ := db.GetFeedByID(feedID)

	// Queue the same feed twice
	go fetcher.FetchSingleFeed(context.Background(), *feed)
	go fetcher.FetchSingleFeed(context.Background(), *feed)

	// Give goroutines time to start and queue
	time.Sleep(100 * time.Millisecond)

	// Check progress - should only have 1 feed queued, not 2
	progress := fetcher.GetProgress()
	if progress.Total > 1 {
		t.Errorf("Expected total 1 (duplicate prevented), got %d", progress.Total)
	}

	// Wait for completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for feed to complete")
		case <-ticker.C:
			progress := fetcher.GetProgress()
			if !progress.IsRunning {
				return
			}
		}
	}
}
