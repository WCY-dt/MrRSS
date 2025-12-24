package models

import "time"

type Feed struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	URL                string    `json:"url"`
	Link               string    `json:"link"` // Website homepage link
	Description        string    `json:"description"`
	Category           string    `json:"category"`
	ImageURL           string    `json:"image_url"` // New field
	Position           int       `json:"position"`  // Position within category for custom ordering
	LastUpdated        time.Time `json:"last_updated"`
	LastError          string    `json:"last_error,omitempty"`  // Track last fetch error
	DiscoveryCompleted bool      `json:"discovery_completed"`   // Track if discovery has been run
	ScriptPath         string    `json:"script_path,omitempty"` // Path to custom script for fetching feed
	HideFromTimeline   bool      `json:"hide_from_timeline"`    // Hide articles from timeline views
	ProxyURL           string    `json:"proxy_url,omitempty"`   // Custom proxy URL for this feed (overrides global)
	ProxyEnabled       bool      `json:"proxy_enabled"`         // Whether to use proxy for this feed
	RefreshInterval    int       `json:"refresh_interval"`      // Custom refresh interval in minutes (0 = use global, -1 = intelligent, >0 = custom minutes)
	IsImageMode        bool      `json:"is_image_mode"`         // Whether this feed is for image gallery mode
	// RSSHub support
	IsRSSHub    bool   `json:"is_rsshub"`              // Whether this feed is from RSSHub
	RSSHubRoute string `json:"rsshub_route,omitempty"` // RSSHub route path (e.g., "/nytimes")
	OriginalURL string `json:"original_url,omitempty"` // Original URL before RSSHub fallback
	// XPath support for HTML/XML scraping
	Type                string `json:"type"`                   // "HTML+XPath" or "XML+XPath"
	XPathItem           string `json:"xpath_item"`             // XPath to extract feed items
	XPathItemTitle      string `json:"xpath_item_title"`       // XPath to extract item title
	XPathItemContent    string `json:"xpath_item_content"`     // XPath to extract item content
	XPathItemUri        string `json:"xpath_item_uri"`         // XPath to extract item URI
	XPathItemAuthor     string `json:"xpath_item_author"`      // XPath to extract item author
	XPathItemTimestamp  string `json:"xpath_item_timestamp"`   // XPath to extract item timestamp
	XPathItemTimeFormat string `json:"xpath_item_time_format"` // Time format for parsing timestamp
	XPathItemThumbnail  string `json:"xpath_item_thumbnail"`   // XPath to extract item thumbnail
	XPathItemCategories string `json:"xpath_item_categories"`  // XPath to extract item categories
	XPathItemUid        string `json:"xpath_item_uid"`         // XPath to extract item unique ID
}

type Article struct {
	ID              int64     `json:"id"`
	FeedID          int64     `json:"feed_id"`
	Title           string    `json:"title"`
	URL             string    `json:"url"`
	ImageURL        string    `json:"image_url"`
	AudioURL        string    `json:"audio_url"`
	VideoURL        string    `json:"video_url"` // YouTube video URL for embedded player
	PublishedAt     time.Time `json:"published_at"`
	IsRead          bool      `json:"is_read"`
	IsFavorite      bool      `json:"is_favorite"`
	IsHidden        bool      `json:"is_hidden"`
	IsReadLater     bool      `json:"is_read_later"`
	FeedTitle       string    `json:"feed_title,omitempty"` // Joined field
	TranslatedTitle string    `json:"translated_title"`
	Summary         string    `json:"summary"` // Cached AI-generated summary
}
