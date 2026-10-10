package factory

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/debug"
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

// APIClient returns a high-level REST client with base URL and auth configured.
func (f *Factory) APIClient() (*api.APIClient, error) {
	actx, err := f.ActiveContext()
	if err != nil {
		return nil, err
	}
	f.debugConfig(actx)
	return api.NewAPIClient(actx.APIURL(), f.newTransport(actx)), nil
}

// ActiveContext returns the context used for API calls. When both INCLOUD_HOST and
// INCLOUD_TOKEN are set it is built from env vars alone, without reading the config file.
func (f *Factory) ActiveContext() (*config.Context, error) {
	if h := os.Getenv("INCLOUD_HOST"); h != "" && config.EnvCredentials() {
		return &config.Context{Host: h}, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}
	return cfg.ActiveContext()
}

func (f *Factory) debugConfig(ctx *config.Context) {
	if !debug.Enabled {
		return
	}

	// Context source
	if os.Getenv("INCLOUD_HOST") != "" && config.EnvCredentials() {
		debug.Log("context: (from: env INCLOUD_HOST + INCLOUD_TOKEN)")
	} else if envCtx := os.Getenv("INCLOUD_CONTEXT"); envCtx != "" {
		debug.Log("context: %s (from: env INCLOUD_CONTEXT)", envCtx)
	} else if cfg, err := f.Config(); err == nil {
		debug.Log("context: %s (from: config)", cfg.CurrentContext)
	}

	// URLs
	debug.Log("api:  %s", ctx.APIURL())
	debug.Log("auth: %s", ctx.AuthURL())

	// Tenant
	if tenant := os.Getenv("INCLOUD_TENANT"); tenant != "" {
		debug.Log("tenant: %s", tenant)
	}

	// User
	if ctx.User != "" {
		debug.Log("user: %s", ctx.User)
	}
}

func (f *Factory) newTransport(ctx *config.Context) *api.TokenTransport {
	t := &api.TokenTransport{
		Token:    ctx.EffectiveToken(),
		APIHost:  ctx.APIURL(),
		AuthHost: ctx.AuthURL(),
		Sudo:     os.Getenv("INCLOUD_SUDO"),
		Tenant:   os.Getenv("INCLOUD_TENANT"),
		Base:     http.DefaultTransport,
	}
	if config.EnvCredentials() {
		return t
	}
	t.RefreshToken = ctx.RefreshToken
	t.OnRefresh = func(accessToken, refreshToken string, expiry time.Time) {
		cfg, err := f.Config()
		if err != nil {
			return
		}
		stored, ok := cfg.Contexts[cfg.ActiveContextName()]
		if !ok {
			return
		}
		stored.Token = accessToken
		if refreshToken != "" {
			stored.RefreshToken = refreshToken
		}
		if !expiry.IsZero() {
			stored.ExpiresAt = expiry
		}
		_ = f.SaveConfig()
	}
	return t
}
