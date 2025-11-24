// Package discovery provides blog discovery functionality
package discovery

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// DiscoveredBlog represents a blog found through friend links
type DiscoveredBlog struct {
	Name           string          `json:"name"`
	Homepage       string          `json:"homepage"`
	RSSFeed        string          `json:"rss_feed"`
	IconURL        string          `json:"icon_url"`
	RecentArticles []RecentArticle `json:"recent_articles"`
}

// RecentArticle represents a recent article with title and date
type RecentArticle struct {
	Title string `json:"title"`
	Date  string `json:"date"` // ISO 8601 format or relative time
}

// Service handles blog discovery operations
type Service struct {
	client     *http.Client
	feedParser *gofeed.Parser
}

// NewService creates a new discovery service
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		feedParser: gofeed.NewParser(),
	}
}

// DiscoverFromFeed discovers blogs from a feed's homepage
func (s *Service) DiscoverFromFeed(ctx context.Context, feedURL string) ([]DiscoveredBlog, error) {
	// First, try to parse the feed to get the homepage link
	homepage, err := s.getFeedHomepage(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get homepage from feed: %w", err)
	}

	// Fetch the homepage HTML
	friendLinks, err := s.findFriendLinks(ctx, homepage)
	if err != nil {
		return nil, fmt.Errorf("failed to find friend links: %w", err)
	}

	if len(friendLinks) == 0 {
		return []DiscoveredBlog{}, nil
	}

	// Discover RSS feeds from friend links (concurrent)
	discovered := s.discoverRSSFeeds(ctx, friendLinks)

	return discovered, nil
}

// getFeedHomepage extracts the homepage URL from a feed
func (s *Service) getFeedHomepage(ctx context.Context, feedURL string) (string, error) {
	feed, err := s.feedParser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return "", err
	}

	if feed.Link != "" {
		return feed.Link, nil
	}

	// Fallback: try to extract base URL from feed URL
	u, err := url.Parse(feedURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// findFriendLinks searches for friend link pages and extracts links
func (s *Service) findFriendLinks(ctx context.Context, homepage string) ([]string, error) {
	// Try to find friend link page
	friendPageURL, err := s.findFriendLinkPage(ctx, homepage)
	if err != nil {
		log.Printf("Could not find friend link page, trying homepage: %v", err)
		friendPageURL = homepage
	}

	// Fetch and parse the friend link page
	doc, err := s.fetchHTML(ctx, friendPageURL)
	if err != nil {
		return nil, err
	}

	// Extract all external links
	links := s.extractExternalLinks(doc, friendPageURL)

	return links, nil
}

// findFriendLinkPage searches for a friend link page
func (s *Service) findFriendLinkPage(ctx context.Context, homepage string) (string, error) {
	doc, err := s.fetchHTML(ctx, homepage)
	if err != nil {
		return "", err
	}

	// Common patterns for friend link pages (multiple languages)
	patterns := []string{
		"友链", "友情链接", "blogroll", "friends", "links",
		"link", "buddy", "partner", "about/links", "friends.html",
	}

	var foundURL string
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		if foundURL != "" {
			return
		}

		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		text := strings.ToLower(strings.TrimSpace(sel.Text()))
		hrefLower := strings.ToLower(href)

		// Check if link text or href contains friend link patterns
		for _, pattern := range patterns {
			if strings.Contains(text, pattern) || strings.Contains(hrefLower, pattern) {
				// Resolve relative URLs
				if absURL := s.resolveURL(homepage, href); absURL != "" {
					foundURL = absURL
					return
				}
			}
		}
	})

	if foundURL != "" {
		return foundURL, nil
	}

	return "", fmt.Errorf("friend link page not found")
}

// fetchHTML fetches and parses HTML from a URL
func (s *Service) fetchHTML(ctx context.Context, urlStr string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "MrRSS/1.0 (Blog Discovery Bot)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// extractExternalLinks extracts all external links from a page
func (s *Service) extractExternalLinks(doc *goquery.Document, baseURL string) []string {
	seen := make(map[string]bool)
	var links []string

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return links
	}

	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		absURL := s.resolveURL(baseURL, href)
		if absURL == "" {
			return
		}

		u, err := url.Parse(absURL)
		if err != nil {
			return
		}

		// Only include external links (different domain)
		if u.Host != baseU.Host && u.Host != "" {
			// Skip common non-blog domains
			if s.isValidBlogDomain(u.Host) && !seen[absURL] {
				seen[absURL] = true
				links = append(links, absURL)
			}
		}
	})

	return links
}

// isValidBlogDomain checks if a domain is likely a blog
func (s *Service) isValidBlogDomain(host string) bool {
	// Skip common non-blog domains
	skipDomains := []string{
		"facebook.com", "twitter.com", "instagram.com", "linkedin.com",
		"youtube.com", "github.com", "stackoverflow.com", "reddit.com",
		"weibo.com", "zhihu.com", "bilibili.com", "douban.com",
		"google.com", "baidu.com", "bing.com", "yahoo.com",
	}

	hostLower := strings.ToLower(host)
	for _, skip := range skipDomains {
		if strings.Contains(hostLower, skip) {
			return false
		}
	}

	return true
}

// resolveURL resolves a relative URL to an absolute URL
func (s *Service) resolveURL(base, href string) string {
	if href == "" {
		return ""
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	return baseURL.ResolveReference(hrefURL).String()
}

// discoverRSSFeeds discovers RSS feeds from a list of blog URLs
func (s *Service) discoverRSSFeeds(ctx context.Context, blogURLs []string) []DiscoveredBlog {
	var wg sync.WaitGroup
	results := make(chan DiscoveredBlog, len(blogURLs))
	sem := make(chan struct{}, 10) // Limit concurrency to 10

	for _, blogURL := range blogURLs {
		select {
		case <-ctx.Done():
			break
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }()

			if blog, err := s.discoverBlogRSS(ctx, u); err == nil {
				results <- blog
			}
		}(blogURL)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var discovered []DiscoveredBlog
	for blog := range results {
		discovered = append(discovered, blog)
	}

	return discovered
}

// discoverBlogRSS discovers RSS feed for a single blog
func (s *Service) discoverBlogRSS(ctx context.Context, blogURL string) (DiscoveredBlog, error) {
	// Try to find RSS feed URL
	rssURL, err := s.findRSSFeed(ctx, blogURL)
	if err != nil {
		return DiscoveredBlog{}, err
	}

	// Parse the RSS feed to get blog info
	feed, err := s.feedParser.ParseURLWithContext(rssURL, ctx)
	if err != nil {
		return DiscoveredBlog{}, err
	}

	// Extract recent articles (max 3)
	var recentArticles []RecentArticle
	for i := 0; i < len(feed.Items) && i < 3; i++ {
		item := feed.Items[i]
		dateStr := ""
		if item.PublishedParsed != nil {
			// Format as relative time or date
			dateStr = item.PublishedParsed.Format("2006-01-02")
		}
		recentArticles = append(recentArticles, RecentArticle{
			Title: item.Title,
			Date:  dateStr,
		})
	}

	// Get favicon
	iconURL := s.getFavicon(blogURL)

	return DiscoveredBlog{
		Name:           feed.Title,
		Homepage:       blogURL,
		RSSFeed:        rssURL,
		IconURL:        iconURL,
		RecentArticles: recentArticles,
	}, nil
}

// findRSSFeed finds the RSS feed URL for a blog
func (s *Service) findRSSFeed(ctx context.Context, blogURL string) (string, error) {
	// Common RSS feed paths to try
	u, err := url.Parse(blogURL)
	if err != nil {
		return "", err
	}

	baseURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	// Common RSS/Atom feed paths
	commonPaths := []string{
		"/rss.xml",
		"/feed.xml",
		"/atom.xml",
		"/feed",
		"/rss",
		"/feeds/posts/default", // Blogger
		"/index.xml",           // Hugo
		"/feed/",
	}

	// Try common paths first
	for _, path := range commonPaths {
		feedURL := baseURL + path
		if s.isValidFeed(ctx, feedURL) {
			return feedURL, nil
		}
	}

	// Try to parse HTML and find RSS link in <head>
	doc, err := s.fetchHTML(ctx, blogURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch blog page: %w", err)
	}

	var foundFeed string
	doc.Find("link[type='application/rss+xml'], link[type='application/atom+xml']").Each(func(i int, sel *goquery.Selection) {
		if foundFeed != "" {
			return
		}
		if href, exists := sel.Attr("href"); exists {
			foundFeed = s.resolveURL(blogURL, href)
		}
	})

	if foundFeed != "" && s.isValidFeed(ctx, foundFeed) {
		return foundFeed, nil
	}

	return "", fmt.Errorf("RSS feed not found")
}

// isValidFeed checks if a URL is a valid RSS/Atom feed
func (s *Service) isValidFeed(ctx context.Context, feedURL string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", feedURL, nil)
	if err != nil {
		return false
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try GET if HEAD doesn't work
		req, err = http.NewRequestWithContext(ctx, "GET", feedURL, nil)
		if err != nil {
			return false
		}

		resp2, err := s.client.Do(req)
		if err != nil {
			return false
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			return false
		}

		// Read first few bytes to check if it's XML
		buf := make([]byte, 512)
		n, err := io.ReadAtLeast(resp2.Body, buf, 1)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return false
		}
		if n == 0 {
			return false
		}
		content := string(buf[:n])

		// Check for XML declaration and RSS/Atom tags
		if strings.Contains(content, "<?xml") ||
			strings.Contains(content, "<rss") ||
			strings.Contains(content, "<feed") ||
			strings.Contains(content, "<atom") {
			return true
		}
		return false
	}

	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "rss") ||
		strings.Contains(contentType, "atom")
}

// getFavicon gets the favicon URL for a blog
func (s *Service) getFavicon(blogURL string) string {
	u, err := url.Parse(blogURL)
	if err != nil {
		return ""
	}

	// Use Google's favicon service as fallback
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", u.Host)
}
