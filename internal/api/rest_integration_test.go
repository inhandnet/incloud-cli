package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAPIClient_Get(t *testing.T) {
	// Record what the server receives
	var gotPath, gotQuery string
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))
	defer srv.Close()

	transport := &TokenTransport{
		Token: "test-token",
		Base:  http.DefaultTransport,
	}
	client := NewAPIClient(srv.URL, transport)

	body, err := client.Get("/api/v1/devices", url.Values{
		"type":  {"cellular"},
		"after": {"2024-01-01"},
		"empty": {""},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify path
	if gotPath != "/api/v1/devices" {
		t.Errorf("expected path /api/v1/devices, got %s", gotPath)
	}

	// Verify empty param was filtered
	q, _ := url.ParseQuery(gotQuery)
	if q.Get("empty") != "" {
		t.Error("empty param should be filtered out")
	}
	if q.Get("type") != "cellular" {
		t.Errorf("expected type=cellular, got %s", q.Get("type"))
	}
	if q.Get("after") != "2024-01-01" {
		t.Errorf("expected after=2024-01-01, got %s", q.Get("after"))
	}

	// Verify auth header
	if gotAuth != "Bearer test-token" {
		t.Errorf("expected Bearer test-token, got %s", gotAuth)
	}

	// Verify body
	if string(body) != `{"result":"ok"}` {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestAPIClient_GetError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, http.DefaultTransport)

	_, err := client.Get("/api/v1/missing", nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if got := err.Error(); got != `HTTP 404: {"error":"not found"}` {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestAPIClient_GetMultiValueParams(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"result":[]}`))
	}))
	defer srv.Close()

	client := NewAPIClient(srv.URL, http.DefaultTransport)

	q := url.Values{}
	q.Add("groups", "aaa")
	q.Add("groups", "bbb")

	_, err := client.Get("/api/v1/test", q)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _ := url.ParseQuery(gotQuery)
	groups := parsed["groups"]
	if len(groups) != 2 || groups[0] != "aaa" || groups[1] != "bbb" {
		t.Errorf("expected groups=[aaa,bbb], got %v", groups)
	}
}
