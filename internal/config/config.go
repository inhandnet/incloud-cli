package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentContext string              `yaml:"current-context"`
	Contexts       map[string]*Context `yaml:"contexts"`
}

type Context struct {
	Host         string    `yaml:"host"`
	Token        string    `yaml:"token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	Org          string    `yaml:"org,omitempty"`
	User         string    `yaml:"user,omitempty"`
	ExpiresAt    time.Time `yaml:"expires_at,omitempty"`
}

// EffectiveToken returns INCLOUD_TOKEN env var if set, else the stored token.
func (c *Context) EffectiveToken() string {
	if t := os.Getenv("INCLOUD_TOKEN"); t != "" {
		return t
	}
	return c.Token
}

// ActiveContext returns the context selected by INCLOUD_CONTEXT env var or current-context field.
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
	return ctx, nil
}

// ActiveContextName returns the resolved context name.
func (cfg *Config) ActiveContextName() string {
	if name := os.Getenv("INCLOUD_CONTEXT"); name != "" {
		return name
	}
	return cfg.CurrentContext
}

func (cfg *Config) SetContext(name string, ctx *Context) {
	if cfg.Contexts == nil {
		cfg.Contexts = make(map[string]*Context)
	}
	cfg.Contexts[name] = ctx
}

func (cfg *Config) DeleteContext(name string) {
	delete(cfg.Contexts, name)
	if cfg.CurrentContext == name {
		cfg.CurrentContext = ""
	}
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

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}
