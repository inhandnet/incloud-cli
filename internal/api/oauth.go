package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	DefaultPort = 18920
)

// FetchClientID retrieves the OAuth client_id from the platform's frontend settings API.
// This is the same endpoint the web Portal uses to get its auth config.
func FetchClientID(ctx context.Context, host string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host+"/api/v1/frontend/settings", http.NoBody)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching frontend settings: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("frontend settings HTTP %d: %s", resp.StatusCode, string(body))
	}

	var settings struct {
		Result struct {
			AuthProvider struct {
				ClientID string `json:"clientId"`
			} `json:"authProvider"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &settings); err != nil {
		return "", fmt.Errorf("parsing frontend settings: %w", err)
	}
	if settings.Result.AuthProvider.ClientID == "" {
		return "", fmt.Errorf("clientId not found in frontend settings")
	}
	return settings.Result.AuthProvider.ClientID, nil
}

// NewOAuthConfig creates an oauth2.Config for the given host and client.
func NewOAuthConfig(host, clientID string, port int) *oauth2.Config {
	return &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  host + "/oauth2/auth",
			TokenURL: host + "/oauth2/token",
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", port),
		Scopes:      []string{"openid", "offline"},
	}
}

// WaitForCallback starts a local HTTP server and waits for the OAuth callback.
// Returns the authorization code.
func WaitForCallback(port int, timeout time.Duration) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error_description")
			if errMsg == "" {
				errMsg = r.URL.Query().Get("error")
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><body><h2>Login failed</h2><p>%s</p></body></html>", html.EscapeString(errMsg)) //nolint:gosec // errMsg is HTML-escaped
			errCh <- fmt.Errorf("OAuth error: %s", errMsg)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h2>Login successful!</h2><p>You can close this tab.</p></body></html>`)
		codeCh <- code
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("failed to start callback server on port %d: %w", port, err)
	}

	go func() { _ = server.Serve(ln) }()
	defer func() { _ = server.Shutdown(context.Background()) }()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("login timed out after %s — no callback received", timeout)
	}
}

// RefreshAccessToken uses the refresh_token to obtain a new access_token.
func RefreshAccessToken(host, clientID, refreshToken string) (*oauth2.Token, error) {
	cfg := &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			TokenURL: host + "/oauth2/token",
		},
	}
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := cfg.TokenSource(context.Background(), token)
	return ts.Token()
}
