package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/inhandnet/incloud-cli/internal/debug"
)

// TokenTransport is an http.RoundTripper that injects Authorization header
// and auto-refreshes tokens on 401 responses.
type TokenTransport struct {
	Token        string
	RefreshToken string
	Host         string
	ClientID     string
	ClientSecret string
	Sudo         string
	OnRefresh    func(accessToken, refreshToken string, expiry time.Time)
	Base         http.RoundTripper
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Token != "" && t.isSameHost(req) {
		req.Header.Set("Authorization", "Bearer "+t.Token)
		if t.Sudo != "" {
			req.Header.Set("Sudo", t.Sudo)
			debug.Log("sudo: %s", t.Sudo)
		}
	}

	resp, err := t.doRoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Auto-refresh on 401
	if resp.StatusCode == 401 && t.RefreshToken != "" {
		resp.Body.Close()
		debug.Log("token expired, refreshing...")

		newToken, err := RefreshAccessToken(t.Host, t.ClientID, t.ClientSecret, t.RefreshToken)
		if err != nil {
			debug.Log("token refresh failed: %v", err)
			return resp, nil // return original 401
		}

		t.Token = newToken.AccessToken
		if newToken.RefreshToken != "" {
			t.RefreshToken = newToken.RefreshToken
		}
		if t.OnRefresh != nil {
			t.OnRefresh(newToken.AccessToken, newToken.RefreshToken, newToken.Expiry)
		}
		debug.Log("token refreshed, new expiry: %s", newToken.Expiry.Format(time.RFC3339))

		// Retry request with new token
		req.Header.Set("Authorization", "Bearer "+t.Token)
		resp, err = t.doRoundTrip(req)
	}
	return resp, err
}

// doRoundTrip executes the request with debug logging for request and response.
func (t *TokenTransport) doRoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	t.debugRequest(req)
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		debug.Log("< request error: %v", err)
		return resp, err
	}
	t.debugResponse(resp, time.Since(start))
	return resp, nil
}

func (t *TokenTransport) debugRequest(req *http.Request) {
	if !debug.Enabled {
		return
	}
	debug.Log("> %s %s", req.Method, req.URL.String())
	for _, h := range []string{"Content-Type", "Authorization", "Sudo"} {
		if v := req.Header.Get(h); v != "" {
			if h == "Authorization" {
				v = "****"
			}
			debug.Log("> %s: %s", h, v)
		}
	}
	if req.Body != nil && req.Body != http.NoBody && req.GetBody != nil {
		if body, err := req.GetBody(); err == nil {
			buf := make([]byte, 4096)
			n, _ := body.Read(buf)
			if n > 0 {
				if n == len(buf) {
					debug.Log("> Body: %s... (truncated)", string(buf[:n]))
				} else {
					debug.Log("> Body: %s", string(buf[:n]))
				}
			}
			body.Close()
		}
	}
}

func (t *TokenTransport) debugResponse(resp *http.Response, elapsed time.Duration) {
	if !debug.Enabled {
		return
	}
	debug.Log("< %d %s (%s)", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed.Round(time.Millisecond))
	for _, h := range []string{"Content-Type", "X-Request-Id", "X-Total-Count"} {
		if v := resp.Header.Get(h); v != "" {
			debug.Log("< %s: %s", h, v)
		}
	}
}

// isSameHost returns true if the request target matches the configured API host,
// preventing auth headers from leaking to third-party domains (e.g. S3 pre-signed URLs).
func (t *TokenTransport) isSameHost(req *http.Request) bool {
	if t.Host == "" {
		return true
	}
	parsed, err := url.Parse(t.Host)
	if err != nil {
		return true
	}
	return req.URL.Host == parsed.Host
}
