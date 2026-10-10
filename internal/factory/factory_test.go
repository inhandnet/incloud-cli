package factory

import (
	"os"
	"path/filepath"
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
	if tr.RefreshToken != "" || tr.OnRefresh != nil {
		t.Error("env-token transport must not refresh")
	}
}
