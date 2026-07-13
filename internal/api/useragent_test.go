package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/build"
)

func TestUserAgentFormat(t *testing.T) {
	ua := UserAgent()
	want := fmt.Sprintf("incloud-cli/%s (%s/%s)", build.Version, runtime.GOOS, runtime.GOARCH)
	if !strings.HasPrefix(ua, want) {
		t.Errorf("UserAgent() = %q, want prefix %q", ua, want)
	}
}

func TestSanitizeClientToken(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"claude-skill/0.2.0", "claude-skill/0.2.0"},
		{"bad value\nwith spaces", "badvaluewithspaces"},
		{"", ""},
		{strings.Repeat("a", 100), strings.Repeat("a", 64)},
	}
	for _, tt := range tests {
		if got := sanitizeClientToken(tt.in); got != tt.want {
			t.Errorf("sanitizeClientToken(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTokenTransportSetsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	client := &http.Client{Transport: &TokenTransport{Base: http.DefaultTransport}}
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotUA != UserAgent() {
		t.Errorf("User-Agent = %q, want %q", gotUA, UserAgent())
	}
}

func TestUserAgentTransportPreservesExisting(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	req.Header.Set("User-Agent", "custom/1.0")
	client := &http.Client{Transport: &UserAgentTransport{}}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotUA != "custom/1.0" {
		t.Errorf("User-Agent = %q, want custom/1.0", gotUA)
	}
}
