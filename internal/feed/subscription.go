package feed

import (
	"MrRSS/internal/models"
	"MrRSS/internal/utils"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/chromedp/chromedp"
	"github.com/mmcdole/gofeed"
)

// sanitizeFeedXML removes or replaces problematic atom:link elements with non-HTTP schemes
// (like file://, javascript:, data:, etc.) that can cause parsing issues.
// This is a workaround for feeds that include local file system links in their XML.
func sanitizeFeedXML(xmlContent string) string {
	// Pattern to match atom:link elements with non-http/https href attributes
	// This handles cases like: <atom:link href="file://..." rel="self" ... />
	pattern := regexp.MustCompile(`<atom:link\s+[^>]*href=["'](file://|javascript:|data:|ftp://)[^"']*["'][^>]*/?>`)

	// Replace all occurrences with empty string (remove the element)
	cleaned := pattern.ReplaceAllString(xmlContent, "")

	// Also handle standalone <link> elements (without atom: prefix)
	linkPattern := regexp.MustCompile(`<link\s+[^>]*href=["'](file://|javascript:|data:|ftp://)[^"']*["'][^>]*/?>`)
	cleaned = linkPattern.ReplaceAllString(cleaned, "")

	utils.DebugLog("sanitizeFeedXML: Removed non-HTTP links from feed XML")
	return cleaned
}

// fetchAndSanitizeFeed fetches feed content and sanitizes it before parsing
func (f *Fetcher) fetchAndSanitizeFeed(ctx context.Context, feedURL string) (string, error) {
	// Use the feed's HTTP client to fetch content
	httpClient, err := f.getHTTPClient(models.Feed{URL: feedURL})
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	xmlContent := string(body)

	// Sanitize the XML to remove problematic links
	cleanedXML := sanitizeFeedXML(xmlContent)

	return cleanedXML, nil
}

// AddSubscription adds a new feed subscription and returns the feed ID.
func (f *Fetcher) AddSubscription(url string, category string, customTitle string) (int64, error) {
	utils.DebugLog("AddSubscription: Starting to add feed from URL: %s", url)

	// Try fetching and sanitizing the feed first
	ctx := context.Background()
	cleanedXML, err := f.fetchAndSanitizeFeed(ctx, url)
	if err != nil {
		utils.DebugLog("AddSubscription: Failed to fetch feed for %s: %v", url, err)
		// Fall through to standard parsing which might handle it differently
	} else {
		// Try parsing the sanitized XML
		parser := gofeed.NewParser()
		// Use the same HTTP client if available (for proxy settings, etc.)
		if gofeedParser, ok := f.fp.(*gofeed.Parser); ok {
			parser.Client = gofeedParser.Client
		}
		parsedFeed, parseErr := parser.ParseString(cleanedXML)
		if parseErr == nil {
			utils.DebugLog("AddSubscription: Successfully parsed sanitized feed for URL: %s", url)
			title := parsedFeed.Title
			if customTitle != "" {
				title = customTitle
			}

			feed := &models.Feed{
				Title:       title,
				URL:         url,
				Link:        parsedFeed.Link,
				Description: parsedFeed.Description,
				Category:    category,
			}

			if parsedFeed.Image != nil {
				feed.ImageURL = parsedFeed.Image.URL
			}

			return f.db.AddFeed(feed)
		}
		utils.DebugLog("AddSubscription: Parsing sanitized feed failed: %v", parseErr)
	}

	// Fallback: Try standard parsing (for backward compatibility)
	utils.DebugLog("AddSubscription: Attempting standard RSS parsing for URL: %s", url)
	parsedFeed, err := f.fp.ParseURL(url)
	if err != nil {
		utils.DebugLog("AddSubscription: Standard RSS parsing failed for %s: %v", url, err)

		// Only attempt JavaScript execution for certain types of errors
		errStr := err.Error()
		shouldTryJS := false

		// Check for errors that typically indicate HTML content instead of XML
		if strings.Contains(errStr, "XML syntax error") ||
			strings.Contains(errStr, "not a valid RSS") ||
			strings.Contains(errStr, "does not contain valid RSS") ||
			strings.Contains(errStr, "expected element type") ||
			strings.Contains(errStr, "Failed to detect feed type") {
			shouldTryJS = true
			utils.DebugLog("AddSubscription: Error indicates HTML content, will attempt JavaScript execution")
		} else {
			utils.DebugLog("AddSubscription: Error does not indicate HTML content, skipping JavaScript execution")
		}

		if shouldTryJS {
			// If standard parsing fails with parsing errors, try executing JavaScript in browser
			utils.DebugLog("AddSubscription: Attempting JavaScript execution for URL: %s", url)
			parsedFeed, err = f.parseFeedWithJavaScript(context.Background(), url, false) // Normal priority for subscription addition
			if err != nil {
				utils.DebugLog("AddSubscription: JavaScript execution also failed: %v", err)
				return 0, fmt.Errorf("both standard parsing and JavaScript execution failed: %w", err)
			}
			utils.DebugLog("AddSubscription: JavaScript execution succeeded")
		} else {
			// For other types of errors (network, etc.), don't try JS execution
			utils.DebugLog("AddSubscription: Returning error without JavaScript execution: %v", err)
			return 0, err
		}
	} else {
		utils.DebugLog("AddSubscription: Standard RSS parsing succeeded for URL: %s", url)
	}

	title := parsedFeed.Title
	if customTitle != "" {
		title = customTitle
	}

	feed := &models.Feed{
		Title:       title,
		URL:         url,
		Link:        parsedFeed.Link,
		Description: parsedFeed.Description,
		Category:    category,
	}

	if parsedFeed.Image != nil {
		feed.ImageURL = parsedFeed.Image.URL
	}

	return f.db.AddFeed(feed)
}

// AddScriptSubscription adds a new feed subscription that uses a custom script
// and returns the feed ID.
func (f *Fetcher) AddScriptSubscription(scriptPath string, category string, customTitle string) (int64, error) {
	// Validate script path
	if f.scriptExecutor == nil {
		return 0, &ScriptError{Message: "script executor not initialized"}
	}

	// Execute script to get initial feed info
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := f.scriptExecutor.ExecuteScript(ctx, scriptPath)
	if err != nil {
		return 0, err
	}

	title := parsedFeed.Title
	if customTitle != "" {
		title = customTitle
	}

	// Use a placeholder URL for script-based feeds
	url := "script://" + scriptPath

	feed := &models.Feed{
		Title:       title,
		URL:         url,
		Link:        parsedFeed.Link,
		Description: parsedFeed.Description,
		Category:    category,
		ScriptPath:  scriptPath,
	}

	if parsedFeed.Image != nil {
		feed.ImageURL = parsedFeed.Image.URL
	}

	return f.db.AddFeed(feed)
}

// AddXPathSubscription adds a new feed subscription that uses XPath expressions
// and returns the feed ID.
func (f *Fetcher) AddXPathSubscription(url string, category string, customTitle string, feedType string, xpathItem string, xpathItemTitle string, xpathItemContent string, xpathItemUri string, xpathItemAuthor string, xpathItemTimestamp string, xpathItemTimeFormat string, xpathItemThumbnail string, xpathItemCategories string, xpathItemUid string) (int64, error) {
	// Validate URL
	if url == "" {
		return 0, &XPathError{
			Operation: "validate",
			Details:   "URL cannot be empty",
		}
	}

	// Validate feed type
	if feedType != "HTML+XPath" && feedType != "XML+XPath" {
		return 0, &XPathError{
			Operation: "validate",
			Details:   fmt.Sprintf("Invalid feed type '%s'. Must be 'HTML+XPath' or 'XML+XPath'", feedType),
		}
	}

	// Validate required XPath item expression
	if xpathItem == "" {
		return 0, &XPathError{
			Operation: "validate",
			Details:   "Item XPath expression is required",
		}
	}

	// Test fetch the URL to ensure it's accessible before adding
	httpClient, err := utils.CreateHTTPClient("", 30*time.Second)
	if err != nil {
		return 0, &XPathError{
			Operation: "fetch",
			URL:       url,
			Details:   "Failed to create HTTP client",
			Err:       err,
		}
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return 0, &XPathError{
			Operation: "fetch",
			URL:       url,
			Details:   "Failed to fetch content. Please check the URL and your network connection",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, &XPathError{
			Operation: "fetch",
			URL:       url,
			Details:   fmt.Sprintf("HTTP error %d: %s. The URL may be invalid or the server may be unreachable", resp.StatusCode, resp.Status),
		}
	}

	// Try to parse the content to ensure XPath expressions work
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, &XPathError{
			Operation: "fetch",
			URL:       url,
			Details:   "Failed to read response body",
			Err:       err,
		}
	}

	// Test parsing based on feed type
	switch feedType {
	case "HTML+XPath":
		doc, err := htmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return 0, &XPathError{
				Operation: "parse",
				URL:       url,
				Details:   "Failed to parse HTML. The page may have invalid HTML or may not be HTML content",
				Err:       err,
			}
		}
		items := htmlquery.Find(doc, xpathItem)
		if len(items) == 0 {
			return 0, &XPathError{
				Operation: "extract",
				URL:       url,
				XPathExpr: xpathItem,
				Details:   "No items found. The Item XPath expression doesn't match any elements. Please check the XPath expression and the page structure",
			}
		}
	case "XML+XPath":
		doc, err := xmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return 0, &XPathError{
				Operation: "parse",
				URL:       url,
				Details:   "Failed to parse XML. The content may not be valid XML",
				Err:       err,
			}
		}
		items := xmlquery.Find(doc, xpathItem)
		if len(items) == 0 {
			return 0, &XPathError{
				Operation: "extract",
				URL:       url,
				XPathExpr: xpathItem,
				Details:   "No items found. The Item XPath expression doesn't match any elements. Please check the XPath expression and the XML structure",
			}
		}
	}

	// All validations passed, create the feed
	title := customTitle
	if title == "" {
		title = "XPath Feed"
	}

	feed := &models.Feed{
		Title:               title,
		URL:                 url,
		Category:            category,
		Type:                feedType,
		XPathItem:           xpathItem,
		XPathItemTitle:      xpathItemTitle,
		XPathItemContent:    xpathItemContent,
		XPathItemUri:        xpathItemUri,
		XPathItemAuthor:     xpathItemAuthor,
		XPathItemTimestamp:  xpathItemTimestamp,
		XPathItemTimeFormat: xpathItemTimeFormat,
		XPathItemThumbnail:  xpathItemThumbnail,
		XPathItemCategories: xpathItemCategories,
		XPathItemUid:        xpathItemUid,
	}

	return f.db.AddFeed(feed)
}

// ImportSubscription imports a feed subscription and returns the feed ID.
func (f *Fetcher) ImportSubscription(title, url, category string) (int64, error) {
	feed := &models.Feed{
		Title:    title,
		URL:      url,
		Link:     "", // Link will be fetched later when feed is refreshed
		Category: category,
	}
	return f.db.AddFeed(feed)
}

// ParseFeed parses an RSS feed from a URL and returns the parsed feed
func (f *Fetcher) ParseFeed(ctx context.Context, url string) (*gofeed.Feed, error) {
	return f.fp.ParseURLWithContext(url, ctx)
}

// ParseFeedWithScript parses an RSS feed, using a custom script or XPath if specified.
// If scriptPath is non-empty, it executes the script.
// If feed.Type is "HTML+XPath" or "XML+XPath", it uses XPath parsing.
// Otherwise, it fetches from the URL as normal.
// priority: true for high-priority requests (like article content fetching), false for normal requests (like feed refresh)
func (f *Fetcher) ParseFeedWithScript(ctx context.Context, url string, scriptPath string, priority bool) (*gofeed.Feed, error) {
	return f.ParseFeedWithFeed(ctx, &models.Feed{URL: url, ScriptPath: scriptPath}, priority)
}

// ParseFeedWithFeed parses a feed using the feed configuration (script or XPath)
func (f *Fetcher) ParseFeedWithFeed(ctx context.Context, feed *models.Feed, priority bool) (*gofeed.Feed, error) {
	// Parse the feed - priority parameter is kept for compatibility but no longer uses priorityMu
	return f.parseFeedWithFeedInternal(ctx, feed, priority)
}

// parseFeedWithFeedInternal does the actual parsing work
func (f *Fetcher) parseFeedWithFeedInternal(ctx context.Context, feed *models.Feed, priority bool) (*gofeed.Feed, error) {
	utils.DebugLog("parseFeedWithFeedInternal: Starting parsing for URL: %s, scriptPath: %s, type: %s, priority: %v", feed.URL, feed.ScriptPath, feed.Type, priority)

	if feed.ScriptPath != "" {
		utils.DebugLog("parseFeedWithFeedInternal: Using script execution for %s", feed.ScriptPath)
		// Execute the custom script to fetch feed
		if f.scriptExecutor == nil {
			return nil, &ScriptError{Message: "Script executor not initialized"}
		}

		// For high priority requests, use shorter timeout
		scriptCtx := ctx
		if priority {
			var cancel context.CancelFunc
			scriptCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
			defer cancel()
		}

		return f.scriptExecutor.ExecuteScript(scriptCtx, feed.ScriptPath)
	}

	// Check if this is an XPath-based feed
	if feed.Type == "HTML+XPath" || feed.Type == "XML+XPath" {
		utils.DebugLog("parseFeedWithFeedInternal: Using XPath parsing for type: %s", feed.Type)
		// For high priority requests, use shorter timeout
		xpathCtx := ctx
		if priority {
			var cancel context.CancelFunc
			xpathCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
			defer cancel()
		}

		return f.parseFeedWithXPath(xpathCtx, feed)
	}

	utils.DebugLog("parseFeedWithFeedInternal: Using traditional URL-based fetching for %s", feed.URL)
	// Use traditional URL-based fetching
	// For high priority requests, use shorter timeout
	fetchCtx := ctx
	if priority {
		var cancel context.CancelFunc
		fetchCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for content fetching
		defer cancel()
	}

	// Try fetching and sanitizing the feed first to handle file:// URLs in atom:link
	utils.DebugLog("parseFeedWithFeedInternal: Attempting to fetch and sanitize feed for %s", feed.URL)
	cleanedXML, sanitizeErr := f.fetchAndSanitizeFeed(fetchCtx, feed.URL)
	if sanitizeErr == nil {
		// Successfully fetched and sanitized, try parsing
		parser := gofeed.NewParser()
		// Use the same HTTP client if available (for proxy settings, etc.)
		if gofeedParser, ok := f.fp.(*gofeed.Parser); ok {
			parser.Client = gofeedParser.Client
		}
		parsedFeed, err := parser.ParseString(cleanedXML)
		if err == nil {
			utils.DebugLog("parseFeedWithFeedInternal: Successfully parsed sanitized feed for %s", feed.URL)
			return parsedFeed, nil
		}
		utils.DebugLog("parseFeedWithFeedInternal: Parsing sanitized feed failed: %v", err)
		// Fall through to standard parsing
	} else {
		utils.DebugLog("parseFeedWithFeedInternal: Sanitization failed: %v", sanitizeErr)
	}

	// Fallback: Try standard parsing first
	utils.DebugLog("parseFeedWithFeedInternal: Attempting standard RSS parsing for %s", feed.URL)
	parsedFeed, err := f.fp.ParseURLWithContext(feed.URL, fetchCtx)
	if err != nil {
		utils.DebugLog("parseFeedWithFeedInternal: Standard RSS parsing failed: %v", err)

		// Only attempt JavaScript execution for certain types of errors that might indicate
		// the content is generated by JavaScript
		errStr := err.Error()
		shouldTryJS := false

		// Check for errors that typically indicate HTML content instead of XML
		if strings.Contains(errStr, "XML syntax error") ||
			strings.Contains(errStr, "not a valid RSS") ||
			strings.Contains(errStr, "does not contain valid RSS") ||
			strings.Contains(errStr, "expected element type") ||
			strings.Contains(errStr, "Failed to detect feed type") {
			shouldTryJS = true
			utils.DebugLog("parseFeedWithFeedInternal: Error indicates HTML content, will attempt JavaScript execution")
		} else {
			utils.DebugLog("parseFeedWithFeedInternal: Error does not indicate HTML content, skipping JavaScript execution")
		}

		if shouldTryJS {
			// If standard parsing fails with parsing errors, try executing JavaScript in browser
			utils.DebugLog("parseFeedWithFeedInternal: Attempting JavaScript execution for %s", feed.URL)
			jsCtx := ctx
			if priority {
				var cancel context.CancelFunc
				jsCtx, cancel = context.WithTimeout(ctx, 15*time.Second) // Shorter timeout for JS execution in high priority
				defer cancel()
			}

			parsedFeed, err = f.parseFeedWithJavaScript(jsCtx, feed.URL, priority)
			if err != nil {
				utils.DebugLog("parseFeedWithFeedInternal: JavaScript execution also failed: %v", err)
				return nil, fmt.Errorf("both standard parsing and JavaScript execution failed: %w", err)
			}
			utils.DebugLog("parseFeedWithFeedInternal: JavaScript execution succeeded")
		} else {
			// For other types of errors (network, etc.), don't try JS execution
			utils.DebugLog("parseFeedWithFeedInternal: Returning error without JavaScript execution: %v", err)
			return nil, err
		}
	} else {
		utils.DebugLog("parseFeedWithFeedInternal: Standard RSS parsing succeeded")
	}

	return parsedFeed, nil
}

// parseFeedWithXPath parses a feed using XPath expressions
func (f *Fetcher) parseFeedWithXPath(_ context.Context, feed *models.Feed) (*gofeed.Feed, error) {
	if feed.XPathItem == "" {
		return nil, &XPathError{
			Operation: "validate",
			Details:   "XPath item expression is required for XPath-based feeds",
		}
	}

	// Fetch the content
	httpClient, err := utils.CreateHTTPClient("", 30*time.Second)
	if err != nil {
		return nil, &XPathError{
			Operation: "fetch",
			URL:       feed.URL,
			Details:   "Failed to create HTTP client",
			Err:       err,
		}
	}
	resp, err := httpClient.Get(feed.URL)
	if err != nil {
		return nil, &XPathError{
			Operation: "fetch",
			URL:       feed.URL,
			Details:   "Failed to fetch content. Please check the URL and your network connection",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &XPathError{
			Operation: "fetch",
			URL:       feed.URL,
			Details:   fmt.Sprintf("HTTP %d: %s. The server may be unreachable or the page may have moved", resp.StatusCode, resp.Status),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &XPathError{
			Operation: "fetch",
			URL:       feed.URL,
			Details:   "Failed to read response body",
			Err:       err,
		}
	}

	// Create gofeed.Feed
	parsedFeed := &gofeed.Feed{
		Title:       feed.Title,
		Link:        feed.URL,
		Description: feed.Description,
		Items:       make([]*gofeed.Item, 0),
	}

	// Parse based on type
	switch feed.Type {
	case "HTML+XPath":
		doc, err := htmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return nil, &XPathError{
				Operation: "parse",
				URL:       feed.URL,
				Details:   "Failed to parse HTML. The page structure may have changed or the content may not be valid HTML",
				Err:       err,
			}
		}
		items := htmlquery.Find(doc, feed.XPathItem)
		if len(items) == 0 {
			return nil, &XPathError{
				Operation: "extract",
				URL:       feed.URL,
				XPathExpr: feed.XPathItem,
				Details:   "No items found. The Item XPath expression doesn't match any elements on the page. The page structure may have changed",
			}
		}

		// Process HTML items
		for _, item := range items {
			gofeedItem := f.extractItemFromHTMLNode(item, feed)
			parsedFeed.Items = append(parsedFeed.Items, gofeedItem)
		}
	case "XML+XPath":
		doc, err := xmlquery.Parse(strings.NewReader(string(body)))
		if err != nil {
			return nil, &XPathError{
				Operation: "parse",
				URL:       feed.URL,
				Details:   "Failed to parse XML. The content may not be valid XML",
				Err:       err,
			}
		}
		items := xmlquery.Find(doc, feed.XPathItem)
		if len(items) == 0 {
			return nil, &XPathError{
				Operation: "extract",
				URL:       feed.URL,
				XPathExpr: feed.XPathItem,
				Details:   "No items found. The Item XPath expression doesn't match any elements in the XML. The structure may have changed",
			}
		}

		// Process XML items
		for _, item := range items {
			gofeedItem := f.extractItemFromXMLNode(item, feed)
			parsedFeed.Items = append(parsedFeed.Items, gofeedItem)
		}
	default:
		return nil, &XPathError{
			Operation: "validate",
			Details:   fmt.Sprintf("Unsupported feed type '%s'. Must be 'HTML+XPath' or 'XML+XPath'", feed.Type),
		}
	}

	return parsedFeed, nil
}

// extractItemFromHTMLNode extracts a gofeed.Item from an HTML node
func (f *Fetcher) extractItemFromHTMLNode(item *html.Node, feed *models.Feed) *gofeed.Item {
	gofeedItem := &gofeed.Item{}

	// Extract title
	if feed.XPathItemTitle != "" {
		if titleNode := htmlquery.FindOne(item, feed.XPathItemTitle); titleNode != nil {
			gofeedItem.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
		}
	}

	// Extract content
	if feed.XPathItemContent != "" {
		if contentNode := htmlquery.FindOne(item, feed.XPathItemContent); contentNode != nil {
			gofeedItem.Content = htmlquery.OutputHTML(contentNode, true)
		}
	}

	// Extract URI
	if feed.XPathItemUri != "" {
		var link string
		// Special handling for @href XPath - get href attribute directly from the item (a tag)
		if feed.XPathItemUri == "./@href" || feed.XPathItemUri == "@href" || feed.XPathItemUri == "href" {
			link = htmlquery.SelectAttr(item, "href")
		} else {
			// For other XPath expressions
			if uriNode := htmlquery.FindOne(item, feed.XPathItemUri); uriNode != nil {
				// Check if this XPath ends with an attribute selector
				if strings.Contains(feed.XPathItemUri, "@") {
					// For attribute XPath expressions, get the text content directly
					link = strings.TrimSpace(htmlquery.InnerText(uriNode))
				} else {
					// Try to get href attribute first (for element nodes)
					if attr := htmlquery.SelectAttr(uriNode, "href"); attr != "" {
						link = attr
					} else {
						// Fallback to inner text for other XPath expressions
						link = strings.TrimSpace(htmlquery.InnerText(uriNode))
					}
				}
			}
		}

		// Additional fallback: if no link found and item is an <a> tag, get href directly
		if link == "" && item != nil && item.Data == "a" {
			link = htmlquery.SelectAttr(item, "href")
		}

		// Resolve relative URLs to absolute URLs
		if link != "" && !strings.HasPrefix(link, "http") {
			baseURL, err := url.Parse(feed.URL)
			if err == nil {
				if ref, err := url.Parse(link); err == nil {
					gofeedItem.Link = baseURL.ResolveReference(ref).String()
				} else {
					gofeedItem.Link = link
				}
			} else {
				gofeedItem.Link = link
			}
		} else {
			gofeedItem.Link = link
		}
	}

	// If no URI was extracted, generate a unique URL for this article
	// This ensures each XPath article has a unique URL to prevent database conflicts
	if gofeedItem.Link == "" {
		// Use feed URL as base and append a hash of the title or content
		uniqueID := gofeedItem.Title
		if uniqueID == "" {
			// Fallback to content or a timestamp-based ID
			if gofeedItem.Content != "" {
				uniqueID = gofeedItem.Content
			} else {
				uniqueID = fmt.Sprintf("xpath-article-%d", time.Now().UnixNano())
			}
		}
		// Create a simple hash of the unique identifier
		hash := fmt.Sprintf("%x", len(uniqueID)) // Simple length-based hash for uniqueness
		gofeedItem.Link = fmt.Sprintf("%s#xpath-%s", feed.URL, hash)
	}

	// Extract author
	if feed.XPathItemAuthor != "" {
		if authorNode := htmlquery.FindOne(item, feed.XPathItemAuthor); authorNode != nil {
			gofeedItem.Author = &gofeed.Person{
				Name: strings.TrimSpace(htmlquery.InnerText(authorNode)),
			}
		}
	}

	// Extract timestamp
	if feed.XPathItemTimestamp != "" {
		if timeNode := htmlquery.FindOne(item, feed.XPathItemTimestamp); timeNode != nil {
			timeStr := strings.TrimSpace(htmlquery.InnerText(timeNode))
			// Remove icon text if present (e.g., "calendar_month 2025-12" -> "2025-12")
			if strings.Contains(timeStr, " ") {
				parts := strings.Split(timeStr, " ")
				// Find the date part (usually the last part that looks like a date)
				for i := len(parts) - 1; i >= 0; i-- {
					part := strings.TrimSpace(parts[i])
					if part != "" && (strings.Contains(part, "-") || strings.Contains(part, "/") || len(part) >= 4) {
						timeStr = part
						break
					}
				}
			}
			if timeStr != "" {
				var parsedTime time.Time
				var err error
				if feed.XPathItemTimeFormat != "" {
					parsedTime, err = time.Parse(feed.XPathItemTimeFormat, timeStr)
				} else {
					// Try common formats
					formats := []string{
						time.RFC3339,
						time.RFC1123,
						"2006-01-02T15:04:05Z07:00",
						"2006-01-02 15:04:05",
						"2006-01-02",
						"2006/01/02",
						"01/02/2006",
						"2006-01",
					}
					for _, format := range formats {
						parsedTime, err = time.Parse(format, timeStr)
						if err == nil {
							break
						}
					}
				}
				if err == nil {
					gofeedItem.PublishedParsed = &parsedTime
				}
			}
		}
	}

	// Extract thumbnail
	if feed.XPathItemThumbnail != "" {
		if thumbNode := htmlquery.FindOne(item, feed.XPathItemThumbnail); thumbNode != nil {
			var imageURL string
			// Check if it's an img src or just text
			if thumbNode.Data == "img" {
				for _, attr := range thumbNode.Attr {
					if attr.Key == "src" {
						imageURL = attr.Val
						break
					}
				}
			} else {
				imageURL = strings.TrimSpace(htmlquery.InnerText(thumbNode))
			}
			// Resolve relative URLs to absolute URLs
			if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(imageURL); err == nil {
						imageURL = baseURL.ResolveReference(ref).String()
					}
				}
			}
			if imageURL != "" {
				gofeedItem.Image = &gofeed.Image{URL: imageURL}
			}
		}
	}

	// Extract categories
	if feed.XPathItemCategories != "" {
		categories := htmlquery.Find(item, feed.XPathItemCategories)
		if len(categories) > 0 {
			gofeedItem.Categories = make([]string, 0, len(categories))
			for _, cat := range categories {
				catText := strings.TrimSpace(htmlquery.InnerText(cat))
				if catText != "" {
					gofeedItem.Categories = append(gofeedItem.Categories, catText)
				}
			}
		}
	}

	// Extract UID
	if feed.XPathItemUid != "" {
		if uidNode := htmlquery.FindOne(item, feed.XPathItemUid); uidNode != nil {
			gofeedItem.GUID = strings.TrimSpace(htmlquery.InnerText(uidNode))
		}
	}

	// If no UID, generate one from link or title
	if gofeedItem.GUID == "" {
		if gofeedItem.Link != "" {
			gofeedItem.GUID = gofeedItem.Link
		} else {
			gofeedItem.GUID = gofeedItem.Title
		}
	}

	return gofeedItem
}

// extractItemFromXMLNode extracts a gofeed.Item from an XML node
func (f *Fetcher) extractItemFromXMLNode(item *xmlquery.Node, feed *models.Feed) *gofeed.Item {
	gofeedItem := &gofeed.Item{}

	// Extract title
	if feed.XPathItemTitle != "" {
		if titleNode := xmlquery.FindOne(item, feed.XPathItemTitle); titleNode != nil {
			gofeedItem.Title = strings.TrimSpace(titleNode.InnerText())
		}
	}

	// Extract content
	if feed.XPathItemContent != "" {
		if contentNode := xmlquery.FindOne(item, feed.XPathItemContent); contentNode != nil {
			gofeedItem.Content = contentNode.OutputXML(true)
		}
	}

	// Extract URI
	if feed.XPathItemUri != "" {
		if uriNode := xmlquery.FindOne(item, feed.XPathItemUri); uriNode != nil {
			link := strings.TrimSpace(uriNode.InnerText())
			// Resolve relative URLs to absolute URLs
			if link != "" && !strings.HasPrefix(link, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(link); err == nil {
						gofeedItem.Link = baseURL.ResolveReference(ref).String()
					} else {
						gofeedItem.Link = link
					}
				} else {
					gofeedItem.Link = link
				}
			} else {
				gofeedItem.Link = link
			}
		}
	}

	// If no URI was extracted, generate a unique URL for this article
	// This ensures each XPath article has a unique URL to prevent database conflicts
	if gofeedItem.Link == "" {
		// Use feed URL as base and append a hash of the title or content
		uniqueID := gofeedItem.Title
		if uniqueID == "" {
			// Fallback to content or a timestamp-based ID
			if gofeedItem.Content != "" {
				uniqueID = gofeedItem.Content
			} else {
				uniqueID = fmt.Sprintf("xpath-article-%d", time.Now().UnixNano())
			}
		}
		// Create a simple hash of the unique identifier
		hash := fmt.Sprintf("%x", len(uniqueID)) // Simple length-based hash for uniqueness
		gofeedItem.Link = fmt.Sprintf("%s#xpath-%s", feed.URL, hash)
	}

	// Extract author
	if feed.XPathItemAuthor != "" {
		if authorNode := xmlquery.FindOne(item, feed.XPathItemAuthor); authorNode != nil {
			gofeedItem.Author = &gofeed.Person{
				Name: strings.TrimSpace(authorNode.InnerText()),
			}
		}
	}

	// Extract timestamp
	if feed.XPathItemTimestamp != "" {
		if timeNode := xmlquery.FindOne(item, feed.XPathItemTimestamp); timeNode != nil {
			timeStr := strings.TrimSpace(timeNode.InnerText())
			if timeStr != "" {
				var parsedTime time.Time
				var err error
				if feed.XPathItemTimeFormat != "" {
					parsedTime, err = time.Parse(feed.XPathItemTimeFormat, timeStr)
				} else {
					// Try common formats
					formats := []string{
						time.RFC3339,
						time.RFC1123,
						"2006-01-02T15:04:05Z07:00",
						"2006-01-02 15:04:05",
						"2006-01-02",
					}
					for _, format := range formats {
						parsedTime, err = time.Parse(format, timeStr)
						if err == nil {
							break
						}
					}
				}
				if err == nil {
					gofeedItem.PublishedParsed = &parsedTime
				}
			}
		}
	}

	// Extract thumbnail
	if feed.XPathItemThumbnail != "" {
		if thumbNode := xmlquery.FindOne(item, feed.XPathItemThumbnail); thumbNode != nil {
			var imageURL string
			// For XML, we assume it's text content or attribute
			if thumbNode.Type == xmlquery.ElementNode && len(thumbNode.Attr) > 0 {
				// Check for src attribute
				for _, attr := range thumbNode.Attr {
					if attr.Name.Local == "src" || attr.Name.Local == "href" {
						imageURL = attr.Value
						break
					}
				}
			} else {
				imageURL = strings.TrimSpace(thumbNode.InnerText())
			}
			// Resolve relative URLs to absolute URLs
			if imageURL != "" && !strings.HasPrefix(imageURL, "http") {
				baseURL, err := url.Parse(feed.URL)
				if err == nil {
					if ref, err := url.Parse(imageURL); err == nil {
						imageURL = baseURL.ResolveReference(ref).String()
					}
				}
			}
			if imageURL != "" {
				gofeedItem.Image = &gofeed.Image{URL: imageURL}
			}
		}
	}

	// Extract categories
	if feed.XPathItemCategories != "" {
		categories := xmlquery.Find(item, feed.XPathItemCategories)
		if len(categories) > 0 {
			gofeedItem.Categories = make([]string, 0, len(categories))
			for _, cat := range categories {
				catText := strings.TrimSpace(cat.InnerText())
				if catText != "" {
					gofeedItem.Categories = append(gofeedItem.Categories, catText)
				}
			}
		}
	}

	// Extract UID
	if feed.XPathItemUid != "" {
		if uidNode := xmlquery.FindOne(item, feed.XPathItemUid); uidNode != nil {
			gofeedItem.GUID = strings.TrimSpace(uidNode.InnerText())
		}
	}

	// If no UID, generate one from link or title
	if gofeedItem.GUID == "" {
		if gofeedItem.Link != "" {
			gofeedItem.GUID = gofeedItem.Link
		} else {
			gofeedItem.GUID = gofeedItem.Title
		}
	}

	return gofeedItem
}

// ScriptError represents an error related to script execution
type ScriptError struct {
	Message string
}

func (e *ScriptError) Error() string {
	return e.Message
}

// XPathError represents an error related to XPath feed operations
type XPathError struct {
	Operation string // "validate", "fetch", "parse", "extract"
	URL       string
	XPathExpr string
	Details   string // Detailed error message
	Err       error  // Underlying error
}

func (e *XPathError) Error() string {
	var msg string
	switch e.Operation {
	case "validate":
		msg = fmt.Sprintf("XPath validation failed for '%s': %s", e.XPathExpr, e.Details)
	case "fetch":
		msg = fmt.Sprintf("Failed to fetch content from %s: %s", e.URL, e.Details)
	case "parse":
		if e.XPathExpr != "" {
			msg = fmt.Sprintf("Failed to parse %s with XPath '%s': %s", e.URL, e.XPathExpr, e.Details)
		} else {
			msg = fmt.Sprintf("Failed to parse %s: %s", e.URL, e.Details)
		}
	case "extract":
		msg = fmt.Sprintf("Failed to extract data with XPath '%s': %s", e.XPathExpr, e.Details)
	default:
		msg = fmt.Sprintf("XPath error: %s", e.Details)
	}

	if e.Err != nil {
		msg += fmt.Sprintf(" (%v)", e.Err)
	}
	return msg
}

func (e *XPathError) Unwrap() error {
	return e.Err
}

// parseFeedWithJavaScript executes JavaScript in the browser and attempts to parse the resulting XML
func (f *Fetcher) parseFeedWithJavaScript(ctx context.Context, feedURL string, priority bool) (*gofeed.Feed, error) {
	utils.DebugLog("parseFeedWithJavaScript: Starting JavaScript execution for URL: %s, priority: %v", feedURL, priority)

	// Create a context with timeout for browser operations
	browserCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Set timeout based on priority
	timeout := 30 * time.Second
	if priority {
		timeout = 15 * time.Second // Shorter timeout for high priority requests
	}

	utils.DebugLog("parseFeedWithJavaScript: Setting timeout to %v", timeout)
	browserCtx, cancel = context.WithTimeout(browserCtx, timeout)
	defer cancel()

	var pageContent string

	// Give some extra time for JavaScript to execute (less for high priority)
	sleepTime := 2 * time.Second
	if priority {
		sleepTime = 1 * time.Second
	}

	utils.DebugLog("parseFeedWithJavaScript: Sleep time set to %v", sleepTime)

	// Run chromedp tasks: navigate to URL and wait for page to load, then get the final HTML
	utils.DebugLog("parseFeedWithJavaScript: Starting chromedp tasks for URL: %s", feedURL)
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(feedURL),
		// Wait for the page to be ready (network idle or DOM content loaded)
		chromedp.WaitReady("body"),
		// Give some extra time for JavaScript to execute
		chromedp.Sleep(sleepTime),
		// Get the final page content
		chromedp.OuterHTML("html", &pageContent),
	)

	if err != nil {
		utils.DebugLog("parseFeedWithJavaScript: chromedp execution failed: %v", err)
		return nil, fmt.Errorf("failed to execute JavaScript in browser: %w", err)
	}

	utils.DebugLog("parseFeedWithJavaScript: chromedp execution succeeded, page content length: %d", len(pageContent))

	// Check if the content is wrapped in Chrome's XML viewer HTML
	if strings.Contains(pageContent, `<div id="webkit-xml-viewer-source-xml">`) {
		utils.DebugLog("parseFeedWithJavaScript: Detected Chrome XML viewer wrapper, extracting XML content")
		// Extract the actual XML content from Chrome's XML viewer
		startTag := `<div id="webkit-xml-viewer-source-xml">`
		endTag := `</div>`
		startIdx := strings.Index(pageContent, startTag)
		if startIdx != -1 {
			startIdx += len(startTag)
			endIdx := strings.Index(pageContent[startIdx:], endTag)
			if endIdx != -1 {
				xmlContent := pageContent[startIdx : startIdx+endIdx]
				utils.DebugLog("parseFeedWithJavaScript: Extracted XML content length: %d", len(xmlContent))
				pageContent = xmlContent
			}
		}
	}

	// Log first 5000 characters of the content for debugging
	if len(pageContent) > 5000 {
		utils.DebugLog("parseFeedWithJavaScript: First 500 chars of page content: %s...", pageContent[:5000])
	} else {
		utils.DebugLog("parseFeedWithJavaScript: Full page content: %s", pageContent)
	}

	// Try to parse the resulting content as RSS/Atom XML
	utils.DebugLog("parseFeedWithJavaScript: Attempting to parse content as RSS/Atom")
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(pageContent)
	if err != nil {
		utils.DebugLog("parseFeedWithJavaScript: RSS/Atom parsing failed: %v", err)
		// Log more details about the content when parsing fails
		utils.DebugLog("parseFeedWithJavaScript: Content type detection - checking for RSS/Atom markers")
		hasRSS := strings.Contains(pageContent, "<rss") || strings.Contains(pageContent, "<feed")
		hasXML := strings.Contains(pageContent, "<?xml")
		utils.DebugLog("parseFeedWithJavaScript: Content analysis - hasRSS: %v, hasXML: %v", hasRSS, hasXML)
		if len(pageContent) > 200 {
			utils.DebugLog("parseFeedWithJavaScript: First 200 chars of failed content: %s", pageContent[:200])
		} else {
			utils.DebugLog("parseFeedWithJavaScript: Full failed content: %s", pageContent)
		}
		return nil, fmt.Errorf("failed to parse content after JavaScript execution: %w", err)
	}

	utils.DebugLog("parseFeedWithJavaScript: RSS/Atom parsing succeeded, feed title: %s, items count: %d", feed.Title, len(feed.Items))
	return feed, nil
}
