package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestFetchClientID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/frontend/settings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"authProvider": map[string]any{
					"clientId":  "test-client-id",
					"authority": "https://auth.example.com/",
				},
			},
		})
	}))
	defer server.Close()

	clientID, err := FetchClientID(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if clientID != "test-client-id" {
		t.Errorf("expected 'test-client-id', got %q", clientID)
	}
}

func TestNewOAuthConfig(t *testing.T) {
	cfg := NewOAuthConfig("https://portal.example.com", "my-client", 18920)
	if cfg.ClientID != "my-client" {
		t.Errorf("unexpected client ID: %s", cfg.ClientID)
	}
	if cfg.RedirectURL != "http://localhost:18920/callback" {
		t.Errorf("unexpected redirect URL: %s", cfg.RedirectURL)
	}
	if cfg.Endpoint.AuthURL != "https://portal.example.com/oauth2/auth" {
		t.Errorf("unexpected auth URL: %s", cfg.Endpoint.AuthURL)
	}
}

func TestWaitForCallback(t *testing.T) {
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Simulate OAuth callback
		http.Get("http://localhost:18921/callback?code=test-auth-code&state=test-state")
	}()
	code, err := WaitForCallback(18921, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if code != "test-auth-code" {
		t.Errorf("expected 'test-auth-code', got %q", code)
	}
}

func TestVerifierIsValid(t *testing.T) {
	// oauth2.GenerateVerifier should produce a valid verifier
	verifier := oauth2.GenerateVerifier()
	if len(verifier) < 43 {
		t.Errorf("verifier too short: %d", len(verifier))
	}
}
