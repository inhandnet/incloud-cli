package factory

import (
	"net/http"
	"sync"
	"time"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type Factory struct {
	IO         *iostreams.IOStreams
	ConfigPath string

	configOnce sync.Once
	config     *config.Config
	configErr  error
}

func New() *Factory {
	return &Factory{
		IO:         iostreams.System(),
		ConfigPath: config.DefaultPath(),
	}
}

func (f *Factory) Config() (*config.Config, error) {
	f.configOnce.Do(func() {
		f.config, f.configErr = config.Load(f.ConfigPath)
	})
	return f.config, f.configErr
}

// ReloadConfig forces config to be reloaded on next access.
func (f *Factory) ReloadConfig() {
	f.configOnce = sync.Once{}
	f.config = nil
	f.configErr = nil
}

func (f *Factory) SaveConfig() error {
	if f.config == nil {
		return nil
	}
	return config.Save(f.config, f.ConfigPath)
}

// HttpClient returns an http.Client with Authorization header injected.
func (f *Factory) HttpClient() (*http.Client, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	ctx, err := cfg.ActiveContext()
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &tokenTransport{
			token:        ctx.EffectiveToken(),
			refreshToken: ctx.RefreshToken,
			host:         ctx.Host,
			clientID:     ctx.ClientID,
			onRefresh: func(accessToken, refreshToken string, expiry time.Time) {
				ctx.Token = accessToken
				if refreshToken != "" {
					ctx.RefreshToken = refreshToken
				}
				if !expiry.IsZero() {
					ctx.ExpiresAt = expiry
				}
				f.SaveConfig()
			},
			base: http.DefaultTransport,
		},
	}, nil
}

type tokenTransport struct {
	token        string
	refreshToken string
	host         string
	clientID     string
	onRefresh    func(accessToken, refreshToken string, expiry time.Time)
	base         http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Auto-refresh on 401
	if resp.StatusCode == 401 && t.refreshToken != "" {
		resp.Body.Close()

		newToken, err := api.RefreshAccessToken(t.host, t.clientID, t.refreshToken)
		if err != nil {
			return resp, nil // return original 401
		}

		t.token = newToken.AccessToken
		if newToken.RefreshToken != "" {
			t.refreshToken = newToken.RefreshToken
		}
		if t.onRefresh != nil {
			t.onRefresh(newToken.AccessToken, newToken.RefreshToken, newToken.Expiry)
		}

		// Retry request with new token
		req.Header.Set("Authorization", "Bearer "+t.token)
		return t.base.RoundTrip(req)
	}
	return resp, err
}
