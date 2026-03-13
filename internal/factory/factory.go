package factory

import (
	"net/http"
	"sync"

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
	token := ctx.EffectiveToken()
	return &http.Client{
		Transport: &tokenTransport{
			token: token,
			base:  http.DefaultTransport,
		},
	}, nil
}

type tokenTransport struct {
	token string
	base  http.RoundTripper
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.base.RoundTrip(req)
}
