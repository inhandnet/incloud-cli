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
	actx, err := f.activeContext()
	if err != nil {
		return nil, err
	}
	f.debugConfig(actx)
	return api.NewAPIClient(actx.APIURL(), f.newTransport(actx)), nil
}

func (f *Factory) activeContext() (*config.Context, error) {
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

	cfg, _ := f.Config()

	// Context source
	if envCtx := os.Getenv("INCLOUD_CONTEXT"); envCtx != "" {
		debug.Log("context: %s (from: env INCLOUD_CONTEXT)", envCtx)
	} else {
		debug.Log("context: %s (from: config)", cfg.CurrentContext)
	}

	// URLs
	debug.Log("api:  %s", ctx.APIURL())
	debug.Log("auth: %s", ctx.AuthURL())

	// Org
	if ctx.Org != "" {
		debug.Log("org: %s", ctx.Org)
	}

	// User
	if ctx.User != "" {
		debug.Log("user: %s", ctx.User)
	}
}

func (f *Factory) newTransport(ctx *config.Context) *api.TokenTransport {
	return &api.TokenTransport{
		Token:        ctx.EffectiveToken(),
		RefreshToken: ctx.RefreshToken,
		APIHost:      ctx.APIURL(),
		AuthHost:     ctx.AuthURL(),
		ClientID:     ctx.ClientID,
		ClientSecret: ctx.ClientSecret,
		Sudo:         os.Getenv("INCLOUD_SUDO"),
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
