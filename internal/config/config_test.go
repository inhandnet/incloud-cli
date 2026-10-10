package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestUpdateAndLoad(t *testing.T) {
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
	if _, err := Update(path, func(c *Config) error { *c = *cfg; return nil }); err != nil {
		t.Fatal(err)
	}

	// verify file permissions
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
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

func TestEnvHostOverrideDoesNotMutateStoredContext(t *testing.T) {
	t.Setenv("INCLOUD_HOST", "https://override.example.com")
	cfg := &Config{
		CurrentContext: "dev",
		Contexts: map[string]*Context{
			"dev": {Host: "https://dev.example.com"},
		},
	}
	if _, err := cfg.ActiveContext(); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Contexts["dev"].Host; got != "https://dev.example.com" {
		t.Errorf("stored context host = %q, want unchanged", got)
	}
}

func TestUpdateReloadsAndKeepsOtherChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if _, err := Update(path, func(c *Config) error {
		c.SetContext("dev", &Context{Host: "https://dev.example.com", Token: "old"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// A process that loaded the config earlier must not write back its stale copy.
	if _, err := Update(path, func(c *Config) error {
		c.Contexts["dev"].Token = "new"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := Update(path, func(c *Config) error {
		c.CurrentContext = "dev"
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Contexts["dev"].Token != "new" || loaded.CurrentContext != "dev" {
		t.Errorf("got token %q, current %q", loaded.Contexts["dev"].Token, loaded.CurrentContext)
	}
}

func TestUpdateErrorLeavesFileUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if _, err := Update(path, func(c *Config) error { c.CurrentContext = "dev"; return nil }); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("boom")
	if _, err := Update(path, func(c *Config) error { c.CurrentContext = "prod"; return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}
	loaded, _ := Load(path)
	if loaded.CurrentContext != "dev" {
		t.Errorf("current context = %q, want dev", loaded.CurrentContext)
	}
}

func TestUpdateConcurrentWritersAndReaders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if _, err := Update(path, func(c *Config) error {
		c.SetContext("dev", &Context{Host: "https://dev.example.com"})
		c.CurrentContext = "dev"
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	const writers = 20
	var wg sync.WaitGroup
	errs := make(chan error, writers*2)
	for i := range writers {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := Update(path, func(c *Config) error {
				c.SetContext(fmt.Sprintf("ctx-%d", i), &Context{Host: "https://example.com"})
				return nil
			})
			errs <- err
		}()
		go func() {
			defer wg.Done()
			cfg, err := Load(path)
			if err == nil {
				if _, err = cfg.ActiveContext(); err != nil {
					err = fmt.Errorf("reader saw incomplete config: %w", err)
				}
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Error(err)
		}
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(loaded.Contexts); got != writers+1 {
		t.Errorf("contexts = %d, want %d (lost updates)", got, writers+1)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file %s", e.Name())
		}
	}
}
