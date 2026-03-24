package auth

import (
	"context"
	"fmt"
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
		Example: `  # Login to dev environment
  incloud auth login --context dev --host nezha.inhand.dev

  # Full URL also works
  incloud auth login --context dev --host https://portal.nezha.inhand.dev

  # Login with explicit client ID (skips auto-detection)
  incloud auth login --context prod --host nezha.inhand.cn --client-id my-client`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.ContextName, "context", "", "Context name to create/update (required)")
	cmd.Flags().StringVar(&opts.Host, "host", "", "Platform host URL (required)")
	cmd.Flags().StringVar(&opts.ClientID, "client-id", "", "OAuth client ID (auto-detected from host if omitted)")
	cmd.Flags().IntVar(&opts.Port, "port", oauthapi.DefaultPort, "Local callback server port")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 2*time.Minute, "Timeout waiting for browser callback")

	_ = cmd.MarkFlagRequired("context")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}

func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("host is required")
	}
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

	if err := validateHost(opts.Host); err != nil {
		return err
	}

	// Derive auth URL from host for OAuth endpoints
	tmpCtx := &config.Context{Host: opts.Host}
	authURL := tmpCtx.AuthURL()

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
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	if !token.Expiry.IsZero() {
		ctx.ExpiresAt = token.Expiry
	}

	cfg.SetContext(opts.ContextName, ctx)
	cfg.CurrentContext = opts.ContextName

	if err := f.SaveConfig(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(out, "%s Logged in to %s (context: %s)\n",
		iostreams.Green("✓"), tmpCtx.APIURL(), opts.ContextName)
	return nil
}
