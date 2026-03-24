package config

import (
	"net"
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

// stripScheme removes scheme and trailing slash from a host string.
func stripScheme(h string) string {
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	return strings.TrimRight(h, "/")
}

// baseDomain extracts the base domain from the Host field by stripping
// scheme, trailing slash, and known service subdomain prefixes.
func (c *Context) baseDomain() string {
	h := stripScheme(c.Host)
	if dot := strings.IndexByte(h, '.'); dot > 0 {
		if knownServicePrefixes[h[:dot]] {
			return h[dot+1:]
		}
	}
	return h
}

// APIURL returns the URL for the API service (star).
// For IP-based hosts (e.g. test servers), returns the original Host unchanged.
func (c *Context) APIURL() string {
	if c.isIPHost() {
		return c.Host
	}
	return "https://star." + c.baseDomain()
}

// AuthURL returns the URL for the auth/portal service.
// For IP-based hosts (e.g. test servers), returns the original Host unchanged.
func (c *Context) AuthURL() string {
	if c.isIPHost() {
		return c.Host
	}
	return "https://portal." + c.baseDomain()
}

// ResolveAPIURL derives the API service URL from a host string.
func ResolveAPIURL(host string) string {
	return (&Context{Host: host}).APIURL()
}

// ResolveAuthURL derives the auth service URL from a host string.
func ResolveAuthURL(host string) string {
	return (&Context{Host: host}).AuthURL()
}

// isIPHost returns true if the Host field points to an IP address
// rather than a domain name (e.g. http://127.0.0.1:8080).
func (c *Context) isIPHost() bool {
	h := stripScheme(c.Host)
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		host = h
	}
	return net.ParseIP(host) != nil
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
