package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CurrentContext != "" {
		t.Errorf("expected empty current context, got %q", cfg.CurrentContext)
	}
	if len(cfg.Contexts) != 0 {
		t.Errorf("expected 0 contexts, got %d", len(cfg.Contexts))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev": {
				Host:  "https://portal.nezha.inhand.dev",
				Token: "tok123",
				User:  "admin",
			},
		},
	}
	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}

	// verify file permissions
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CurrentContext != "dev" {
		t.Errorf("expected current context 'dev', got %q", loaded.CurrentContext)
	}
	ctx, ok := loaded.Contexts["dev"]
	if !ok {
		t.Fatal("context 'dev' not found")
	}
	if ctx.Host != "https://portal.nezha.inhand.dev" {
		t.Errorf("unexpected host: %s", ctx.Host)
	}
	if ctx.Token != "tok123" {
		t.Errorf("unexpected token: %s", ctx.Token)
	}
}

func TestSetAndDeleteContext(t *testing.T) {
	cfg := &Config{Contexts: make(map[string]*Context)}

	cfg.SetContext("prod", &Context{Host: "https://prod.example.com", User: "admin"})
	if _, ok := cfg.Contexts["prod"]; !ok {
		t.Fatal("context not set")
	}

	cfg.DeleteContext("prod")
	if _, ok := cfg.Contexts["prod"]; ok {
		t.Fatal("context not deleted")
	}
}

func TestCurrentContextObj(t *testing.T) {
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev": {Host: "https://dev.example.com"},
		},
	}
	ctx, err := cfg.ActiveContext()
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Host != "https://dev.example.com" {
		t.Errorf("unexpected host: %s", ctx.Host)
	}

	cfg.CurrentContext = "nonexistent"
	_, err = cfg.ActiveContext()
	if err == nil {
		t.Fatal("expected error for nonexistent context")
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("INCLOUD_TOKEN", "env-token")
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev": {Host: "https://dev.example.com", Token: "file-token"},
		},
	}
	ctx, _ := cfg.ActiveContext()
	token := ctx.EffectiveToken()
	if token != "env-token" {
		t.Errorf("expected env token override, got %q", token)
	}
}

func TestEnvHostOverride(t *testing.T) {
	t.Setenv("INCLOUD_HOST", "https://override.example.com")
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev": {Host: "https://dev.example.com"},
		},
	}
	ctx, _ := cfg.ActiveContext()
	if ctx.Host != "https://override.example.com" {
		t.Errorf("expected INCLOUD_HOST override, got %q", ctx.Host)
	}
}

func TestEnvContextOverride(t *testing.T) {
	t.Setenv("INCLOUD_CONTEXT", "prod")
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev":  {Host: "https://dev.example.com"},
			"prod": {Host: "https://prod.example.com"},
		},
	}
	ctx, _ := cfg.ActiveContext()
	if ctx.Host != "https://prod.example.com" {
		t.Errorf("expected INCLOUD_CONTEXT to select prod, got %q", ctx.Host)
	}
	if cfg.ActiveContextName() != "prod" {
		t.Errorf("expected ActiveContextName 'prod', got %q", cfg.ActiveContextName())
	}
}
