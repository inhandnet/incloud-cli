package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

// HTTPError represents an HTTP response with a non-2xx status code.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, string(e.Body))
}

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

// UploadWithFields performs a multipart file upload via POST with additional form fields.
func (c *APIClient) UploadWithFields(path, fieldName, fileName string, reader io.Reader, fields map[string]string) ([]byte, error) {
	r := c.inner.R().SetFileReader(fieldName, fileName, reader)
	if len(fields) > 0 {
		r.SetFormData(fields)
	}
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
		return body, &HTTPError{StatusCode: resp.StatusCode(), Body: body}
	}
	return body, nil
}

// Download performs a GET request and writes the response body to a file.
// Suitable for binary downloads (e.g. pcap files, firmware images).
// path can be a relative API path or an absolute URL.
func (c *APIClient) Download(path, destFile string) error {
	reqURL := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		reqURL = c.inner.BaseURL + path
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	resp, err := c.inner.GetClient().Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	f, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(destFile)
		return fmt.Errorf("writing file: %w", err)
	}
	return f.Close()
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
