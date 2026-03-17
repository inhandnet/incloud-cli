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
	actx, err := f.activeContext()
	if err != nil {
		return nil, err
	}
	return &http.Client{Transport: f.newTransport(actx)}, nil
}

// APIClient returns a high-level REST client with base URL and auth configured.
// Use this for the common GET+parse pattern; use HttpClient() for advanced cases.
func (f *Factory) APIClient() (*api.APIClient, error) {
	actx, err := f.activeContext()
	if err != nil {
		return nil, err
	}
	return api.NewAPIClient(actx.Host, f.newTransport(actx)), nil
}

func (f *Factory) activeContext() (*config.Context, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	return cfg.ActiveContext()
}

func (f *Factory) newTransport(ctx *config.Context) *api.TokenTransport {
	return &api.TokenTransport{
		Token:        ctx.EffectiveToken(),
		RefreshToken: ctx.RefreshToken,
		Host:         ctx.Host,
		ClientID:     ctx.ClientID,
		ClientSecret: ctx.ClientSecret,
		OnRefresh: func(accessToken, refreshToken string, expiry time.Time) {
			ctx.Token = accessToken
			if refreshToken != "" {
				ctx.RefreshToken = refreshToken
			}
			if !expiry.IsZero() {
				ctx.ExpiresAt = expiry
			}
			_ = f.SaveConfig()
		},
		Base: http.DefaultTransport,
	}
}
