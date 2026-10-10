package factory

import (
	"context"
	"errors"
	"fmt"
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

// UpdateConfig applies fn to the latest config on disk under a file lock and saves it.
// fn must only change the fields it owns. The in-memory config is replaced by the result.
func (f *Factory) UpdateConfig(fn func(*config.Config) error) error {
	cfg, err := config.Update(f.ConfigPath, fn)
	if err != nil {
		return err
	}
	f.configOnce.Do(func() {})
	f.config, f.configErr = cfg, nil
	return nil
}

// refreshTimeout bounds a token refresh, which holds the config lock while it runs.
const refreshTimeout = 20 * time.Second

// RefreshToken refreshes the token of context name under the config lock and returns the
// updated context. If the token on disk is no longer staleToken, another process has
// already refreshed it and that token is returned without refreshing again.
func (f *Factory) RefreshToken(ctx context.Context, name, authHost, staleToken string) (*config.Context, error) {
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	var refreshed config.Context
	err := f.UpdateConfig(func(cfg *config.Config) error {
		stored, ok := cfg.Contexts[name]
		if !ok {
			return fmt.Errorf("context %q not found", name)
		}
		if stored.Token != "" && stored.Token != staleToken {
			refreshed = *stored
			return nil
		}
		if stored.RefreshToken == "" {
			return errors.New("no refresh token")
		}
		token, err := api.RefreshToken(ctx, authHost, stored.RefreshToken)
		if err != nil {
			return err
		}
		stored.Token = token.AccessToken
		if token.RefreshToken != "" {
			stored.RefreshToken = token.RefreshToken
		}
		if !token.Expiry.IsZero() {
			stored.ExpiresAt = token.Expiry
		}
		refreshed = *stored
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &refreshed, nil
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
		Token:   ctx.EffectiveToken(),
		APIHost: ctx.APIURL(),
		Sudo:    os.Getenv("INCLOUD_SUDO"),
		Tenant:  os.Getenv("INCLOUD_TENANT"),
		Base:    http.DefaultTransport,
	}
	if config.EnvCredentials() || ctx.RefreshToken == "" {
		return t
	}
	cfg, err := f.Config()
	if err != nil {
		return t
	}
	name, authHost := cfg.ActiveContextName(), ctx.AuthURL()
	t.Refresh = func(rctx context.Context, staleToken string) (string, error) {
		refreshed, err := f.RefreshToken(rctx, name, authHost, staleToken)
		if err != nil {
			return "", err
		}
		return refreshed.Token, nil
	}
	return t
}
