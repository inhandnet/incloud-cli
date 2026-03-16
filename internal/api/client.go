package api

import (
	"net/http"
	"time"
)

// TokenTransport is an http.RoundTripper that injects Authorization header
// and auto-refreshes tokens on 401 responses.
type TokenTransport struct {
	Token        string
	RefreshToken string
	Host         string
	ClientID     string
	ClientSecret string
	OnRefresh    func(accessToken, refreshToken string, expiry time.Time)
	Base         http.RoundTripper
}

func (t *TokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Auto-refresh on 401
	if resp.StatusCode == 401 && t.RefreshToken != "" {
		resp.Body.Close()

		newToken, err := RefreshAccessToken(t.Host, t.ClientID, t.ClientSecret, t.RefreshToken)
		if err != nil {
			return resp, nil // return original 401
		}

		t.Token = newToken.AccessToken
		if newToken.RefreshToken != "" {
			t.RefreshToken = newToken.RefreshToken
		}
		if t.OnRefresh != nil {
			t.OnRefresh(newToken.AccessToken, newToken.RefreshToken, newToken.Expiry)
		}

		// Retry request with new token
		req.Header.Set("Authorization", "Bearer "+t.Token)
		return t.Base.RoundTrip(req)
	}
	return resp, err
}
