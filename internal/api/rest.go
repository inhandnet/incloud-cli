package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

// APIClient is a high-level HTTP client with base URL and auth pre-configured.
// It wraps resty to provide a concise, consistent API for CLI commands.
//
// All methods return (responseBody, error). On HTTP >= 400, the error contains
// the status code and response body; the body is also returned so callers can
// inspect it if needed.
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
	return c.execute(r, resty.MethodGet, path)
}

// Post performs a POST request. body is JSON-marshaled by resty; pass nil for
// an empty body.
func (c *APIClient) Post(path string, body interface{}) ([]byte, error) {
	r := c.inner.R()
	if body != nil {
		r.SetBody(body)
	}
	return c.execute(r, resty.MethodPost, path)
}

// Put performs a PUT request. body is JSON-marshaled by resty; pass nil for
// an empty body.
func (c *APIClient) Put(path string, body interface{}) ([]byte, error) {
	r := c.inner.R()
	if body != nil {
		r.SetBody(body)
	}
	return c.execute(r, resty.MethodPut, path)
}

// Delete performs a DELETE request.
func (c *APIClient) Delete(path string) ([]byte, error) {
	return c.execute(c.inner.R(), resty.MethodDelete, path)
}

// Upload performs a multipart file upload via POST.
func (c *APIClient) Upload(path, fieldName, fileName string, reader io.Reader) ([]byte, error) {
	r := c.inner.R().SetFileReader(fieldName, fileName, reader)
	return c.execute(r, resty.MethodPost, path)
}

// RequestOptions configures a generic request via Do.
type RequestOptions struct {
	Query       url.Values
	Body        interface{}
	RawBody     io.Reader
	Headers     map[string]string
	ContentType string
}

// Do performs an arbitrary HTTP request. Use this for the generic `api` command
// or any case that doesn't fit the CRUD helpers.
func (c *APIClient) Do(method, path string, opts *RequestOptions) ([]byte, error) {
	r := c.inner.R()
	if opts != nil {
		if clean := cleanValues(opts.Query); len(clean) > 0 {
			r.SetQueryParamsFromValues(clean)
		}
		if len(opts.Headers) > 0 {
			r.SetHeaders(opts.Headers)
		}
		if opts.RawBody != nil {
			r.SetBody(opts.RawBody)
			if opts.ContentType != "" {
				r.SetHeader("Content-Type", opts.ContentType)
			}
		} else if opts.Body != nil {
			r.SetBody(opts.Body)
		}
	}
	return c.execute(r, method, path)
}

// execute is the shared request execution + error handling.
func (c *APIClient) execute(r *resty.Request, method, path string) ([]byte, error) {
	resp, err := r.Execute(method, path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	body := resp.Body()
	if resp.IsError() {
		return body, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), string(body))
	}
	return body, nil
}

// HTTPClient returns the underlying *http.Client with auth transport configured.
// Use this for SSE streaming or other cases that need raw http.Client access.
func (c *APIClient) HTTPClient() *http.Client {
	return c.inner.GetClient()
}

// BaseURL returns the configured base URL.
func (c *APIClient) BaseURL() string {
	return c.inner.BaseURL
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
