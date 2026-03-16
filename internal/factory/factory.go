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
		Transport: &api.TokenTransport{
			Token:        ctx.EffectiveToken(),
			RefreshToken: ctx.RefreshToken,
			Host:         ctx.Host,
			ClientID:     ctx.ClientID,
			OnRefresh: func(accessToken, refreshToken string, expiry time.Time) {
				ctx.Token = accessToken
				if refreshToken != "" {
					ctx.RefreshToken = refreshToken
				}
				if !expiry.IsZero() {
					ctx.ExpiresAt = expiry
				}
				f.SaveConfig()
			},
			Base: http.DefaultTransport,
		},
	}, nil
}
