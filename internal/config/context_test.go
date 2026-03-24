package config

import "testing"

func TestContext_APIURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"full portal URL", "https://portal.nezha.inhand.dev", "https://star.nezha.inhand.dev"},
		{"full star URL", "https://star.nezha.inhand.dev", "https://star.nezha.inhand.dev"},
		{"bare domain", "nezha.inhand.dev", "https://star.nezha.inhand.dev"},
		{"custom domain", "https://custom.company.com", "https://star.custom.company.com"},
		{"bare custom", "company.com", "https://star.company.com"},
		{"trailing slash", "https://portal.nezha.inhand.dev/", "https://star.nezha.inhand.dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{Host: tt.host}
			if got := c.APIURL(); got != tt.want {
				t.Errorf("APIURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContext_AuthURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{"full portal URL", "https://portal.nezha.inhand.dev", "https://portal.nezha.inhand.dev"},
		{"full star URL", "https://star.nezha.inhand.dev", "https://portal.nezha.inhand.dev"},
		{"bare domain", "nezha.inhand.dev", "https://portal.nezha.inhand.dev"},
		{"custom domain", "https://custom.company.com", "https://portal.custom.company.com"},
		{"bare custom", "company.com", "https://portal.company.com"},
		{"trailing slash", "https://portal.nezha.inhand.dev/", "https://portal.nezha.inhand.dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{Host: tt.host}
			if got := c.AuthURL(); got != tt.want {
				t.Errorf("AuthURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
