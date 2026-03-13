package auth

import (
	"context"
	"fmt"
	"time"

	oauthapi "github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
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
  incloud auth login --context dev --host https://portal.nezha.inhand.dev

  # Login with explicit client ID (skips auto-detection)
  incloud auth login --context prod --host https://portal.nezha.inhand.cn --client-id my-client`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Host == "" {
				return fmt.Errorf("--host is required")
			}
			if opts.ContextName == "" {
				return fmt.Errorf("--context is required")
			}
			return runLogin(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.ContextName, "context", "", "Context name to create/update (required)")
	cmd.Flags().StringVar(&opts.Host, "host", "", "Platform host URL (required)")
	cmd.Flags().StringVar(&opts.ClientID, "client-id", "", "OAuth client ID (auto-detected from host if omitted)")
	cmd.Flags().IntVar(&opts.Port, "port", oauthapi.DefaultPort, "Local callback server port")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 2*time.Minute, "Timeout waiting for browser callback")

	return cmd
}

func runLogin(f *factory.Factory, opts *LoginOptions) error {
	out := f.IO.Out

	// 1. Resolve client_id: from flag or auto-detect from host
	clientID := opts.ClientID
	if clientID == "" {
		fmt.Fprintln(out, "Fetching OAuth client configuration...")
		var err error
		clientID, err = oauthapi.FetchClientID(opts.Host)
		if err != nil {
			return fmt.Errorf("auto-detecting client_id from %s: %w\n  Hint: use --client-id to specify manually", opts.Host, err)
		}
		fmt.Fprintf(out, "Using client: %s\n", iostreams.Gray(clientID))
	}

	// 2. Build OAuth config and generate PKCE verifier
	oauthCfg := oauthapi.NewOAuthConfig(opts.Host, clientID, opts.Port)
	verifier := oauth2.GenerateVerifier()
	state := "incloud-cli-login"

	// 3. Build auth URL with PKCE and open browser
	authURL := oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
	)

	fmt.Fprintln(out, "Opening browser for authentication...")
	fmt.Fprintln(out, iostreams.Gray("If the browser doesn't open, visit:"))
	fmt.Fprintln(out, iostreams.Gray(authURL))
	fmt.Fprintln(out)

	if err := browser.OpenURL(authURL); err != nil {
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
		iostreams.Green("✓"), opts.Host, opts.ContextName)
	return nil
}
