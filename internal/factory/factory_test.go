package factory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/config"
)

func newTestFactory(t *testing.T) *Factory {
	t.Helper()
	return &Factory{ConfigPath: filepath.Join(t.TempDir(), "incloud", "config.yaml")}
}

func TestActiveContext_EnvHostAndTokenNeedNoConfigFile(t *testing.T) {
	t.Setenv("INCLOUD_HOST", "https://portal.example.com")
	t.Setenv("INCLOUD_TOKEN", "env-token")
	t.Setenv("INCLOUD_CONTEXT", "missing")
	f := newTestFactory(t)

	ctx, err := f.ActiveContext()
	if err != nil {
		t.Fatalf("ActiveContext() error = %v", err)
	}
	if ctx.Host != "https://portal.example.com" {
		t.Errorf("Host = %q", ctx.Host)
	}
	if _, err := f.APIClient(); err != nil {
		t.Fatalf("APIClient() error = %v", err)
	}
	if _, err := os.Stat(filepath.Dir(f.ConfigPath)); !os.IsNotExist(err) {
		t.Errorf("config dir exists after env-mode calls (stat err = %v)", err)
	}
}

func TestNewTransport_EnvTokenDisablesRefresh(t *testing.T) {
	t.Setenv("INCLOUD_TOKEN", "env-token")
	f := newTestFactory(t)

	tr := f.newTransport(&config.Context{Host: "https://portal.example.com", RefreshToken: "file-refresh"})
	if tr.Token != "env-token" {
		t.Errorf("Token = %q, want env-token", tr.Token)
	}
	if tr.Refresh != nil {
		t.Error("env-token transport must not refresh")
	}
}

func seedConfig(t *testing.T, f *Factory, ctx *config.Context) {
	t.Helper()
	if err := f.UpdateConfig(func(c *config.Config) error {
		c.SetContext("dev", ctx)
		c.CurrentContext = "dev"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func newRefreshServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var refreshes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/frontend/settings":
			_, _ = w.Write([]byte(`{"result":{"authProvider":{"clientId":"id","clientSecret":"secret"}}}`))
		case "/oauth2/token":
			refreshes.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fresh","refresh_token":"fresh-refresh","token_type":"Bearer","expires_in":3600}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &refreshes
}

func TestRefreshToken_RefreshesAndPersists(t *testing.T) {
	t.Setenv("INCLOUD_TOKEN", "")
	srv, refreshes := newRefreshServer(t)
	f := newTestFactory(t)
	seedConfig(t, f, &config.Context{Host: srv.URL, Token: "stale", RefreshToken: "old-refresh"})

	got, err := f.RefreshToken(context.Background(), "dev", srv.URL, "stale")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if got.Token != "fresh" || refreshes.Load() != 1 {
		t.Errorf("token = %q, refreshes = %d", got.Token, refreshes.Load())
	}
	loaded, _ := config.Load(f.ConfigPath)
	if c := loaded.Contexts["dev"]; c.Token != "fresh" || c.RefreshToken != "fresh-refresh" || c.ExpiresAt.IsZero() {
		t.Errorf("persisted context = %+v", c)
	}
}

func TestRefreshToken_ReusesTokenRefreshedByAnotherProcess(t *testing.T) {
	t.Setenv("INCLOUD_TOKEN", "")
	srv, refreshes := newRefreshServer(t)
	f := newTestFactory(t)
	seedConfig(t, f, &config.Context{Host: srv.URL, Token: "stale", RefreshToken: "old-refresh"})

	// Another process refreshes after this one loaded the config.
	if _, err := config.Update(f.ConfigPath, func(c *config.Config) error {
		c.Contexts["dev"].Token = "other"
		c.Contexts["dev"].RefreshToken = "other-refresh"
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	got, err := f.RefreshToken(context.Background(), "dev", srv.URL, "stale")
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if got.Token != "other" || refreshes.Load() != 0 {
		t.Errorf("token = %q, refreshes = %d, want other and 0", got.Token, refreshes.Load())
	}
	loaded, _ := config.Load(f.ConfigPath)
	if loaded.Contexts["dev"].RefreshToken != "other-refresh" {
		t.Errorf("refresh token overwritten: %q", loaded.Contexts["dev"].RefreshToken)
	}
}

func TestAPIClient_PersistsRefreshedToken(t *testing.T) {
	t.Setenv("INCLOUD_TOKEN", "")
	t.Setenv("INCLOUD_HOST", "")
	var apiCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/frontend/settings":
			_, _ = w.Write([]byte(`{"result":{"authProvider":{"clientId":"id","clientSecret":"secret"}}}`))
		case "/oauth2/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fresh","token_type":"Bearer","expires_in":3600}`))
		default:
			if apiCalls.Add(1) == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"result":{}}`))
		}
	}))
	t.Cleanup(srv.Close)
	f := newTestFactory(t)
	seedConfig(t, f, &config.Context{Host: srv.URL, Token: "stale", RefreshToken: "refresh"})

	client, err := f.APIClient()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("/api/v1/users/me", nil); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	loaded, _ := config.Load(f.ConfigPath)
	if loaded.Contexts["dev"].Token != "fresh" {
		t.Errorf("persisted token = %q, want fresh", loaded.Contexts["dev"].Token)
	}
}
