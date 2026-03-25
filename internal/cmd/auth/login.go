package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	oauthapi "github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type LoginOptions struct {
	ContextName string
	Host        string
	ClientID    string // optional override; if empty, fetched from host
	Port        int
	Timeout     time.Duration
}

func NewCmdLogin(f *factory.Factory) *cobra.Command {
	opts := &LoginOptions{}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login via browser OAuth flow",
		Example: `  # Quick start — zero config, logs into global region
  incloud login

  # Login to China region
  incloud login --host cn

  # Use a named context for multi-environment setups
  incloud login --context prod --host global
  incloud login --context cn --host cn

  # Custom domain also works
  incloud login --host inhandcloud.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.ContextName, "context", "default", "Context name to create/update")
	cmd.Flags().StringVar(&opts.Host, "host", "global", `Platform host or region: "global", "cn", or a custom domain`)
	cmd.Flags().StringVar(&opts.ClientID, "client-id", "", "OAuth client ID (auto-detected from host if omitted)")
	cmd.Flags().IntVar(&opts.Port, "port", oauthapi.DefaultPort, "Local callback server port")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 2*time.Minute, "Timeout waiting for browser callback")

	return cmd
}

// regionHosts maps short region names to platform domains.
var regionHosts = map[string]string{
	"global": "inhandcloud.com",
	"cn":     "inhandcloud.cn",
	"dev":    "nezha.inhand.dev",
	"beta":   "nezha.inhand.design",
}

// resolveHost expands region short names (e.g. "cn" → "inhandcloud.cn")
// and validates the result.
func resolveHost(host string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("host is required")
	}
	if domain, ok := regionHosts[strings.ToLower(host)]; ok {
		return domain, nil
	}
	return host, validateHost(host)
}

func validateHost(host string) error {
	// Bare domain (no scheme) — valid
	if !strings.Contains(host, "://") {
		return nil
	}
	// Full URL — validate scheme and no path
	u, err := url.Parse(host)
	if err != nil || u.Host == "" {
		return fmt.Errorf("invalid host URL %q", host)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid host URL %q: scheme must be http or https", host)
	}
	if path := strings.TrimSuffix(u.Path, "/"); path != "" {
		return fmt.Errorf("invalid host URL %q: should not include a path, use %s://%s instead", host, u.Scheme, u.Host)
	}
	return nil
}

func runLogin(f *factory.Factory, opts *LoginOptions) error {
	out := f.IO.Out

	host, err := resolveHost(opts.Host)
	if err != nil {
		return err
	}
	opts.Host = host

	// Derive auth URL from host for OAuth endpoints
	authURL := config.ResolveAuthURL(opts.Host)

	// 1. Resolve client credentials: from flag or auto-detect from host
	clientID := opts.ClientID
	var clientSecret string
	if clientID == "" {
		fmt.Fprintln(out, "Fetching OAuth client configuration...")
		client, err := oauthapi.FetchOAuthClient(context.Background(), authURL)
		if err != nil {
			return fmt.Errorf("auto-detecting client from %s: %w\n  Hint: use --client-id to specify manually", authURL, err)
		}
		clientID = client.ClientID
		clientSecret = client.ClientSecret
		fmt.Fprintf(out, "Using client: %s\n", iostreams.Gray(clientID))
	}

	// 2. Build OAuth config and generate PKCE verifier
	oauthCfg := oauthapi.NewOAuthConfig(authURL, clientID, clientSecret, opts.Port)
	verifier := oauth2.GenerateVerifier()
	state := "incloud-cli-login"

	// 3. Build auth URL with PKCE and open browser
	authorizeURL := oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
	)

	fmt.Fprintln(out, "Opening browser for authentication...")
	fmt.Fprintln(out, iostreams.Gray("If the browser doesn't open, visit:"))
	fmt.Fprintln(out, iostreams.Gray(authorizeURL))
	fmt.Fprintln(out)

	if err := browser.OpenURL(authorizeURL); err != nil {
		fmt.Fprintln(out, iostreams.Yellow("Failed to open browser automatically."))
	}

	// 4. Wait for callback
	fmt.Fprintln(out, "Waiting for authentication...")
	code, err := oauthapi.WaitForCallback(opts.Port, opts.Timeout)
	if err != nil {
		return err
	}

	// 5. Exchange code for token using x/oauth2 with PKCE verifier
	fmt.Fprintln(out, "Exchanging authorization code...")
	token, err := oauthCfg.Exchange(context.Background(), code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// 6. Save to config
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	ctx := &config.Context{
		Host:         opts.Host,
		Token:        token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		ctx.ExpiresAt = token.Expiry
	}

	// Fetch current user identity for display and config storage
	if user := fetchCurrentUser(ctx); user != "" {
		ctx.User = user
	}

	cfg.SetContext(opts.ContextName, ctx)
	cfg.CurrentContext = opts.ContextName

	if err := f.SaveConfig(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	loginMsg := fmt.Sprintf("%s Logged in to %s (context: %s)",
		iostreams.Green("✓"), config.ResolveAPIURL(opts.Host), opts.ContextName)
	if ctx.User != "" {
		loginMsg += fmt.Sprintf(" as %s", iostreams.Bold(ctx.User))
	}
	fmt.Fprintln(out, loginMsg)
	return nil
}

// fetchCurrentUser calls /api/v1/users/me to get the logged-in user's display name.
// Returns "username (email)" or just "username". On any error, returns "".
func fetchCurrentUser(ctx *config.Context) string {
	apiURL := ctx.APIURL()
	transport := &oauthapi.TokenTransport{
		Token: ctx.Token,
		Base:  http.DefaultTransport,
	}
	client := oauthapi.NewAPIClient(apiURL, transport)
	q := url.Values{}
	q.Set("fields", "username,email")
	body, err := client.Get("/api/v1/users/me", q)
	if err != nil {
		return ""
	}
	var resp struct {
		Result struct {
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"result"`
	}
	if json.Unmarshal(body, &resp) != nil {
		return ""
	}
	if resp.Result.Username == "" {
		return ""
	}
	if resp.Result.Email != "" {
		return fmt.Sprintf("%s (%s)", resp.Result.Username, resp.Result.Email)
	}
	return resp.Result.Username
}
