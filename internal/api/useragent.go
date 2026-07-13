package api

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/inhandnet/incloud-cli/internal/build"
)

var userAgentOnce = sync.OnceValue(func() string {
	ua := fmt.Sprintf("incloud-cli/%s (%s/%s)", build.Version, runtime.GOOS, runtime.GOARCH)
	if client := sanitizeClientToken(os.Getenv("INCLOUD_CLIENT")); client != "" {
		ua += " " + client
	}
	return ua
})

// UserAgent returns the User-Agent header value for all outgoing requests,
// e.g. "incloud-cli/1.2.3 (darwin/arm64)". If the INCLOUD_CLIENT environment
// variable is set (e.g. "claude-skill/0.2.0" when invoked by an AI agent via
// the incloud skill), it is appended as an extra product token so server-side
// logs can distinguish agent-driven usage from humans.
func UserAgent() string {
	return userAgentOnce()
}

// sanitizeClientToken keeps only characters valid in a User-Agent product
// token and caps the length, so a malformed env value can't break the header.
func sanitizeClientToken(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '.' || r == '-' || r == '_' || r == '/':
			return r
		}
		return -1
	}, s)
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

// UserAgentTransport is an http.RoundTripper that sets the User-Agent header
// on requests that don't already have one. Use it for HTTP clients that don't
// go through TokenTransport (e.g. the oauth2 token client).
type UserAgentTransport struct {
	Base http.RoundTripper
}

func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", UserAgent())
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
