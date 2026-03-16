package config

import (
	"os"
	"time"
)

type Context struct {
	Host         string    `yaml:"host"`
	Token        string    `yaml:"token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	ClientID     string    `yaml:"client_id,omitempty"`
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
