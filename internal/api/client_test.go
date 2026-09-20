package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestTokenTransport_StripsAuthOnCrossOriginRedirect(t *testing.T) {
	// Simulate: API at star.example.com redirects to s3.amazonaws.com
	transport := &TokenTransport{
		Token:    "secret-token",
		APIHost:  "https://star.example.com",
		AuthHost: "https://portal.example.com",
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "star.example.com" {
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

	ctx := context.Background()

	// Same-host request — should have auth
	sameHost, _ := http.NewRequestWithContext(ctx, "GET", "https://star.example.com/api/v1/files", http.NoBody)
	resp, _ := transport.RoundTrip(sameHost)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}

	// Cross-origin request (simulating a redirect to S3) — should NOT have auth
	crossOrigin, _ := http.NewRequestWithContext(ctx, "GET", "https://s3.amazonaws.com/bucket/file?X-Amz-Algorithm=AWS4", http.NoBody)
	resp, _ = transport.RoundTrip(crossOrigin)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func TestTokenTransport_RetriesWithOriginalBodyAfterTokenRefresh(t *testing.T) {
	const requestBody = `{"name":"edge-1","enabled":true}`

	var (
		mu        sync.Mutex
		apiBodies []string
	)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/devices":
			if r.ProtoMajor != 2 {
				t.Errorf("API request used HTTP/%d; test requires HTTP/2", r.ProtoMajor)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request body: %v", err)
				return
			}
			mu.Lock()
			apiBodies = append(apiBodies, string(body))
			requestCount := len(apiBodies)
			mu.Unlock()

			if requestCount == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Authorization") != "Bearer refreshed-token" {
				t.Errorf("retry used Authorization %q", r.Header.Get("Authorization"))
			}
			w.WriteHeader(http.StatusOK)
		case "/api/v1/frontend/settings":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":{"authProvider":{"clientId":"client-id","clientSecret":"client-secret"}}}`))
		case "/oauth2/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"refreshed-token","token_type":"Bearer","expires_in":3600}`))
		default:
			http.NotFound(w, r)
		}
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	previousDefaultClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = previousDefaultClient }()

	transport := &TokenTransport{
		Token:        "expired-token",
		RefreshToken: "refresh-token",
		APIHost:      server.URL,
		AuthHost:     server.URL,
		Base:         server.Client().Transport,
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/devices", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("RoundTrip() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(apiBodies) != 2 {
		t.Fatalf("API request count = %d, want 2", len(apiBodies))
	}
	for attempt, body := range apiBodies {
		if body != requestBody {
			t.Errorf("API request %d body = %q, want %q", attempt+1, body, requestBody)
		}
	}
}

func TestTokenTransport_RetriesRequestWithoutBodyAfterTokenRefresh(t *testing.T) {
	authServer := newTokenRefreshServer(t)
	attempts := 0
	transport := &TokenTransport{
		Token:        "expired-token",
		RefreshToken: "refresh-token",
		AuthHost:     authServer.URL,
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			if req.Body != nil && req.Body != http.NoBody {
				t.Error("retry request unexpectedly has a body")
			}
			if attempts == 1 {
				return response(http.StatusUnauthorized), nil
			}
			if req.Header.Get("Authorization") != "Bearer refreshed-token" {
				t.Errorf("retry used Authorization %q", req.Header.Get("Authorization"))
			}
			return response(http.StatusOK), nil
		}),
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1/devices", http.NoBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("RoundTrip() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if attempts != 2 {
		t.Errorf("request count = %d, want 2", attempts)
	}
}

func TestTokenTransport_DoesNotRetryNonReplayableBodyAfterTokenRefresh(t *testing.T) {
	authServer := newTokenRefreshServer(t)
	attempts := 0
	transport := &TokenTransport{
		Token:        "expired-token",
		RefreshToken: "refresh-token",
		AuthHost:     authServer.URL,
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			_, _ = io.Copy(io.Discard, req.Body)
			return response(http.StatusUnauthorized), nil
		}),
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.example.com/v1/devices", io.NopCloser(strings.NewReader(`{"name":"edge-1"}`)))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.GetBody = nil

	resp, err := transport.RoundTrip(req)
	if resp != nil {
		resp.Body.Close()
		t.Fatalf("RoundTrip() response = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("RoundTrip() error = nil, want a non-replayable body error")
	}
	if attempts != 1 {
		t.Errorf("request count = %d, want 1", attempts)
	}
}

func TestTokenTransport_ReturnsInitialUnauthorizedResponseWhenRefreshFails(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer authServer.Close()

	previousDefaultClient := http.DefaultClient
	http.DefaultClient = authServer.Client()
	defer func() { http.DefaultClient = previousDefaultClient }()

	attempts := 0
	transport := &TokenTransport{
		Token:        "expired-token",
		RefreshToken: "refresh-token",
		AuthHost:     authServer.URL,
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			return response(http.StatusUnauthorized), nil
		}),
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1/devices", http.NoBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v, want nil", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("RoundTrip() status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if attempts != 1 {
		t.Errorf("request count = %d, want 1", attempts)
	}
}

func TestTokenTransport_DoesNotRefreshOnNonUnauthorizedResponse(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("attempted to refresh after a non-401 response")
	}))
	defer authServer.Close()

	previousDefaultClient := http.DefaultClient
	http.DefaultClient = authServer.Client()
	defer func() { http.DefaultClient = previousDefaultClient }()

	attempts := 0
	transport := &TokenTransport{
		Token:        "current-token",
		RefreshToken: "refresh-token",
		AuthHost:     authServer.URL,
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			attempts++
			return response(http.StatusForbidden), nil
		}),
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1/devices", http.NoBody)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("RoundTrip() status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	if attempts != 1 {
		t.Errorf("request count = %d, want 1", attempts)
	}
}

func TestTokenTransport_EmptyHost_AlwaysAddsAuth(t *testing.T) {
	transport := &TokenTransport{
		Token:   "secret-token",
		APIHost: "",
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") == "" {
				t.Error("expected Authorization header when Host is empty")
			}
			return &http.Response{StatusCode: 200}, nil
		}),
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://anywhere.example.com/path", http.NoBody)
	resp, _ := transport.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
}

func newTokenRefreshServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/frontend/settings":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":{"authProvider":{"clientId":"client-id","clientSecret":"client-secret"}}}`))
		case "/oauth2/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"refreshed-token","token_type":"Bearer","expires_in":3600}`))
		default:
			http.NotFound(w, r)
		}
	}))

	previousDefaultClient := http.DefaultClient
	http.DefaultClient = server.Client()
	t.Cleanup(func() {
		http.DefaultClient = previousDefaultClient
		server.Close()
	})
	return server
}

func response(statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
