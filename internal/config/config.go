package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentContext string              `yaml:"current-context"`
	Contexts       map[string]*Context `yaml:"contexts"`
}

// ActiveContext returns a copy of the context selected by INCLOUD_CONTEXT env var
// or current-context field. If INCLOUD_HOST is set, it overrides the copy's Host field.
func (cfg *Config) ActiveContext() (*Context, error) {
	name := os.Getenv("INCLOUD_CONTEXT")
	if name == "" {
		name = cfg.CurrentContext
	}
	if name == "" {
		return nil, fmt.Errorf("no active context; run 'incloud auth login' or 'incloud config use-context <name>'")
	}
	ctx, ok := cfg.Contexts[name]
	if !ok {
		return nil, fmt.Errorf("context %q not found in config", name)
	}
	c := *ctx
	if h := os.Getenv("INCLOUD_HOST"); h != "" {
		c.Host = h
	}
	return &c, nil
}

// ActiveContextName returns the resolved context name.
func (cfg *Config) ActiveContextName() string {
	if name := os.Getenv("INCLOUD_CONTEXT"); name != "" {
		return name
	}
	return cfg.CurrentContext
}

// DefaultPath returns ~/.config/incloud/config.yaml
func DefaultPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "incloud", "config.yaml")
}

func Load(path string) (*Config, error) {
	cfg := &Config{Contexts: make(map[string]*Context)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]*Context)
	}
	return cfg, nil
}

// lockTimeout bounds how long Update waits for another process holding the config lock.
const lockTimeout = 30 * time.Second

// Update locks the config file, reloads it from disk, applies fn and writes the result
// atomically. fn must only change the fields it owns, so concurrent processes don't
// overwrite each other's changes. Returns the config as written.
func Update(path string, fn func(*Config) error) (*Config, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}

	lock := flock.New(path + ".lock")
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	locked, err := lock.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("locking config: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("locking config: timed out")
	}
	defer func() { _ = lock.Unlock() }()

	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	if err := fn(cfg); err != nil {
		return nil, err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}
	if err := writeAtomic(path, data); err != nil {
		return nil, fmt.Errorf("writing config: %w", err)
	}
	return cfg, nil
}

// writeAtomic writes data to a temp file in the same directory and renames it over path,
// so readers without the lock always see a complete file.
func writeAtomic(path string, data []byte) error {
	// CreateTemp creates the file with 0600 permissions.
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := writeAndClose(tmp, data); err != nil {
		_ = os.Remove(tmpName) //nolint:gosec // tmpName is a temp file next to the config file
		return err
	}
	if err := renameWithRetry(tmpName, path); err != nil {
		_ = os.Remove(tmpName) //nolint:gosec // tmpName is a temp file next to the config file
		return err
	}
	return nil
}

func writeAndClose(f *os.File, data []byte) error {
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

// renameWithRetry retries on Windows, where rename fails while another process has the target open.
func renameWithRetry(from, to string) error {
	var err error
	for range 10 {
		if err = os.Rename(from, to); err == nil || runtime.GOOS != "windows" { //nolint:gosec // paths are the config file and its temp file
			return err
		}
		time.Sleep(20 * time.Millisecond)
	}
	return err
}
