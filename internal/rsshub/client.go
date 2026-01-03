package rsshub

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client handles RSSHub route validation and URL transformation
type Client struct {
	Endpoint string
	APIKey   string
}

// NewClient creates a new RSSHub client
func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		APIKey:   apiKey,
	}
}

// ValidateRoute performs a HEAD request to check if route exists
func (c *Client) ValidateRoute(route string) error {
	url := c.BuildURL(route)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.APIKey != "" {
		q := req.URL.Query()
		q.Add("key", c.APIKey)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to RSSHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("route not found: %s", route)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("RSSHub returned error: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// BuildURL converts a route to full RSSHub URL
func (c *Client) BuildURL(route string) string {
	url := fmt.Sprintf("%s/%s", c.Endpoint, route)
	if c.APIKey != "" {
		url = fmt.Sprintf("%s?key=%s", url, c.APIKey)
	}
	return url
}

// IsRSSHubURL checks if a URL uses the rsshub:// protocol
func IsRSSHubURL(url string) bool {
	return strings.HasPrefix(url, "rsshub://")
}

// ExtractRoute removes the rsshub:// prefix
func ExtractRoute(url string) string {
	return strings.TrimPrefix(url, "rsshub://")
}
