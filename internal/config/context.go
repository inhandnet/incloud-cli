package config

import (
	"os"
	"strings"
	"time"
)

type Context struct {
	Host         string    `yaml:"host"`
	Token        string    `yaml:"token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	ClientID     string    `yaml:"client_id,omitempty"`
	ClientSecret string    `yaml:"client_secret,omitempty"`
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

// knownServicePrefixes are subdomain prefixes that represent platform
// services and should be stripped when deriving the base domain.
var knownServicePrefixes = map[string]bool{
	"portal": true,
	"star":   true,
}

// baseDomain extracts the base domain from the Host field by stripping
// scheme, trailing slash, and known service subdomain prefixes.
func (c *Context) baseDomain() string {
	h := c.Host
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	h = strings.TrimRight(h, "/")

	if dot := strings.IndexByte(h, '.'); dot > 0 {
		prefix := h[:dot]
		if knownServicePrefixes[prefix] {
			return h[dot+1:]
		}
	}
	return h
}

// APIURL returns the URL for the API service (star).
func (c *Context) APIURL() string {
	return "https://star." + c.baseDomain()
}

// AuthURL returns the URL for the auth/portal service.
func (c *Context) AuthURL() string {
	return "https://portal." + c.baseDomain()
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
