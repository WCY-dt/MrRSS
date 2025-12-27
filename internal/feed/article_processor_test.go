package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

func TestExtractMediaThumbnail(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url":    "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
												"width":  "480",
												"height": "360",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
		},
		{
			name: "Direct media:thumbnail structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"thumbnail": []ext.Extension{
							{
								Name:  "thumbnail",
								Value: "",
								Attrs: map[string]string{
									"url": "https://example.com/thumb.jpg",
								},
							},
						},
					},
				},
			},
			expected: "https://example.com/thumb.jpg",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
		{
			name: "Media extensions without thumbnail",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"title": []ext.Extension{
							{
								Name:  "title",
								Value: "Some Title",
							},
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaThumbnail(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaThumbnail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractMediaTitle(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"title": {
										{
											Name:  "title",
											Value: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
		},
		{
			name: "Direct media:title structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"title": []ext.Extension{
							{
								Name:  "title",
								Value: "Direct Media Title",
							},
						},
					},
				},
			},
			expected: "Direct Media Title",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaTitle(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractMediaDescription(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"description": {
										{
											Name:  "description",
											Value: "I'm joining bush pilot Matt Dearden as we fly into extreme airstrips...",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "I'm joining bush pilot Matt Dearden as we fly into extreme airstrips...",
		},
		{
			name: "Direct media:description structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"description": []ext.Extension{
							{
								Name:  "description",
								Value: "Direct Media Description",
							},
						},
					},
				},
			},
			expected: "Direct Media Description",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaDescription(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaDescription() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractImageURLWithMediaRSS(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:thumbnail",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url": "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
		},
		{
			name: "Item with Image takes precedence",
			item: &gofeed.Item{
				Image: &gofeed.Image{
					URL: "https://example.com/item-image.jpg",
				},
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url": "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://example.com/item-image.jpg",
		},
		{
			name: "Fallback to enclosure",
			item: &gofeed.Item{
				Enclosures: []*gofeed.Enclosure{
					{
						URL:  "https://example.com/enclosure-image.png",
						Type: "image/png",
					},
				},
			},
			expected: "https://example.com/enclosure-image.png",
		},
		{
			name: "Fallback to HTML img tag",
			item: &gofeed.Item{
				Description: `<p>Some text</p><img src="https://example.com/html-image.jpg" alt="test">`,
			},
			expected: "https://example.com/html-image.jpg",
		},
		{
			name:     "No image available",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractImageURL(tt.item)
			if result != tt.expected {
				t.Errorf("extractImageURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessArticlesWithYouTubeFeed(t *testing.T) {
	// Create a mock database
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Create a mock fetcher
	f := &Fetcher{
		db: db,
		// translator will be nil for this test
	}

	// Create a mock feed
	feed := models.Feed{
		ID: 1,
	}

	// Create mock YouTube feed items
	publishedTime := time.Now()
	items := []*gofeed.Item{
		{
			Title: "WORST Place to be a Pilot",
			Link:  "https://www.youtube.com/watch?v=KZcE7HgtFsA",
			Extensions: ext.Extensions{
				"media": {
					"group": []ext.Extension{
						{
							Name:  "group",
							Value: "",
							Children: map[string][]ext.Extension{
								"title": {
									{
										Name:  "title",
										Value: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
									},
								},
								"description": {
									{
										Name:  "description",
										Value: "I'm joining bush pilot Matt Dearden as we fly into some of the world's most extreme and unforgiving airstrips.",
									},
								},
								"thumbnail": {
									{
										Name:  "thumbnail",
										Value: "",
										Attrs: map[string]string{
											"url":    "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											"width":  "480",
											"height": "360",
										},
									},
								},
							},
						},
					},
				},
			},
			PublishedParsed: &publishedTime,
		},
	}

	// Process the articles
	articlesWithContent := f.processArticles(feed, items)

	// Verify results
	if len(articlesWithContent) != 1 {
		t.Fatalf("Expected 1 article, got %d", len(articlesWithContent))
	}

	article := articlesWithContent[0].Article

	// Should use the longer media:title
	expectedTitle := "WORST Place to be a Pilot: West Papua's Extreme Bush Flying"
	if article.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, article.Title)
	}

	// Should extract media:thumbnail
	expectedImageURL := "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg"
	if article.ImageURL != expectedImageURL {
		t.Errorf("Expected image URL '%s', got '%s'", expectedImageURL, article.ImageURL)
	}

	// Should have correct URL
	expectedURL := "https://www.youtube.com/watch?v=KZcE7HgtFsA"
	if article.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, article.URL)
	}

	// Should extract video URL for embedded player
	expectedVideoURL := "https://www.youtube.com/embed/KZcE7HgtFsA"
	if article.VideoURL != expectedVideoURL {
		t.Errorf("Expected video URL '%s', got '%s'", expectedVideoURL, article.VideoURL)
	}
}
