package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

// APIClient is a high-level HTTP client with base URL and auth pre-configured.
// It wraps resty to provide a concise API for the common GET+parse pattern
// used across CLI commands.
type APIClient struct {
	inner *resty.Client
}

// NewAPIClient creates an APIClient with the given base URL and transport.
// The transport typically carries auth (e.g. TokenTransport).
func NewAPIClient(baseURL string, transport http.RoundTripper) *APIClient {
	c := resty.New()
	c.SetBaseURL(baseURL)
	c.SetTransport(transport)
	return &APIClient{inner: c}
}

// Get performs a GET request. Empty values in query params are automatically
// skipped so callers don't need conditional checks.
func (c *APIClient) Get(path string, query url.Values) ([]byte, error) {
	r := c.inner.R()
	if clean := cleanValues(query); len(clean) > 0 {
		r.SetQueryParamsFromValues(clean)
	}
	resp, err := r.Get(path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	body := resp.Body()
	if resp.IsError() {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(body))
	}
	return body, nil
}

// cleanValues returns a copy of v with empty string values removed.
func cleanValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	clean := make(url.Values)
	for k, vals := range v {
		for _, val := range vals {
			if val != "" {
				clean.Add(k, val)
			}
		}
	}
	return clean
}
