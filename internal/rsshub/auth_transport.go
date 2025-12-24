package rsshub

import (
	"net/http"
)

type AuthTransport struct {
	baseHTTP http.RoundTripper
	apiKey   string
}

func NewAuthTransport(apiKey string) *AuthTransport {
	return &AuthTransport{
		baseHTTP: http.DefaultTransport,
		apiKey:   apiKey,
	}
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	req = req.Clone(req.Context())

	if t.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.apiKey)
	}

	// Call the base transport
	return t.baseHTTP.RoundTrip(req)
}

// SetBaseTransport sets the underlying HTTP transport
func (t *AuthTransport) SetBaseTransport(base http.RoundTripper) {
	t.baseHTTP = base
}
