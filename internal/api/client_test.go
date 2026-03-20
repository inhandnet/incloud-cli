package api

import (
	"net/http"
	"testing"
)

func TestTokenTransport_StripsAuthOnCrossOriginRedirect(t *testing.T) {
	// Simulate: API at portal.example.com redirects to s3.amazonaws.com
	transport := &TokenTransport{
		Token: "secret-token",
		Host:  "https://portal.example.com",
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "portal.example.com" {
				// Expect auth header on same-host request
				if req.Header.Get("Authorization") == "" {
					t.Error("expected Authorization header for same-host request")
				}
			}
			if req.URL.Host == "s3.amazonaws.com" {
				// Must NOT have auth header on cross-origin request
				if req.Header.Get("Authorization") != "" {
					t.Error("Authorization header leaked to cross-origin host")
				}
			}
			return &http.Response{StatusCode: 200}, nil
		}),
	}

	// Same-host request — should have auth
	sameHost, _ := http.NewRequest("GET", "https://portal.example.com/api/v1/files", nil)
	transport.RoundTrip(sameHost)

	// Cross-origin request (simulating a redirect to S3) — should NOT have auth
	crossOrigin, _ := http.NewRequest("GET", "https://s3.amazonaws.com/bucket/file?X-Amz-Algorithm=AWS4", nil)
	transport.RoundTrip(crossOrigin)
}

func TestTokenTransport_EmptyHost_AlwaysAddsAuth(t *testing.T) {
	transport := &TokenTransport{
		Token: "secret-token",
		Host:  "",
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") == "" {
				t.Error("expected Authorization header when Host is empty")
			}
			return &http.Response{StatusCode: 200}, nil
		}),
	}

	req, _ := http.NewRequest("GET", "https://anywhere.example.com/path", nil)
	transport.RoundTrip(req)
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
