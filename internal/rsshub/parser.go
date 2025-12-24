package rsshub

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseRSSHubURL parses a complete RSSHub URL into its components
// Returns the route path and query parameters
// Note: The 'key' parameter is excluded as it will be added from settings automatically
func ParseRSSHubURL(rssHubURL string) (string, map[string]string, error) {
	parsedURL, err := url.Parse(rssHubURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Get route path
	routePath := parsedURL.Path

	// Parse query parameters (exclude 'key' as it's managed in settings)
	queryParams := make(map[string]string)
	for key, values := range parsedURL.Query() {
		// Skip the 'key' parameter - it will be added from settings automatically
		if key == "key" {
			continue
		}
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	return routePath, queryParams, nil
}

// BuildRSSHubURL builds a complete RSSHub URL from components
// Example: instanceURL="https://rsshub.app", route="/weibo/user/123", params={limit: "10"}, apiKey="key"
// Returns: https://rsshub.app/weibo/user/123?limit=10&key=key
func BuildRSSHubURL(instanceURL string, route string, params map[string]string, apiKey string) string {
	// Normalize instance URL
	instanceURL = strings.TrimSuffix(instanceURL, "/")

	// Build query string from params
	queryString := ""
	if len(params) > 0 {
		parts := make([]string, 0, len(params))
		for key, value := range params {
			parts = append(parts, fmt.Sprintf("%s=%s", key, url.QueryEscape(value)))
		}
		queryString = strings.Join(parts, "&")
	}

	rssHubURL := instanceURL + route
	if queryString != "" {
		rssHubURL += "?" + queryString
	}

	// Add API key if provided
	if apiKey != "" {
		separator := "?"
		if strings.Contains(rssHubURL, "?") {
			separator = "&"
		}
		rssHubURL += separator + "key=" + url.QueryEscape(apiKey)
	}

	return rssHubURL
}

// ExtractInstanceURL extracts the instance URL from a complete RSSHub URL
// Example: https://rsshub.app/weibo/user/123 -> https://rsshub.app
func ExtractInstanceURL(rssHubURL string) (string, error) {
	parsedURL, err := url.Parse(rssHubURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Return scheme + host
	return parsedURL.Scheme + "://" + parsedURL.Host, nil
}

// MatchesInstanceURL checks if a feed URL matches a configured RSSHub instance URL
func MatchesInstanceURL(feedURL, instanceURL string) bool {
	feedParsed, err := url.Parse(feedURL)
	if err != nil {
		return false
	}

	instanceParsed, err := url.Parse(instanceURL)
	if err != nil {
		return false
	}

	// Compare hosts
	return feedParsed.Host == instanceParsed.Host
}
