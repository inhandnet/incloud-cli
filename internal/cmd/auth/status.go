package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name := cfg.ActiveContextName()
			if name == "" {
				fmt.Fprintln(f.IO.Out, "No active context")
				return nil
			}

			ctx, ok := cfg.Contexts[name]
			if !ok {
				return fmt.Errorf("context %q not found", name)
			}

			out := f.IO.Out
			fmt.Fprintf(out, "Context:  %s\n", iostreams.Bold(name))
			fmt.Fprintf(out, "Host:     %s\n", ctx.Host)

			if ctx.User != "" {
				fmt.Fprintf(out, "User:     %s\n", ctx.User)
			}
			if ctx.Org != "" {
				fmt.Fprintf(out, "Org:      %s\n", ctx.Org)
			}

			tokenExpired := !ctx.ExpiresAt.IsZero() && ctx.ExpiresAt.Before(time.Now())

			// Try auto-refresh if token expired but refresh token is available
			if tokenExpired && ctx.RefreshToken != "" && ctx.ClientID != "" {
				newToken, err := api.RefreshAccessToken(ctx.Host, ctx.ClientID, ctx.ClientSecret, ctx.RefreshToken)
				if err == nil {
					ctx.Token = newToken.AccessToken
					if newToken.RefreshToken != "" {
						ctx.RefreshToken = newToken.RefreshToken
					}
					if !newToken.Expiry.IsZero() {
						ctx.ExpiresAt = newToken.Expiry
					}
					_ = f.SaveConfig()
					tokenExpired = false
					fmt.Fprintf(out, "Status:   %s\n", iostreams.Green("logged in (token refreshed)"))
				} else {
					fmt.Fprintf(out, "Status:   %s\n", iostreams.Red("token expired, refresh failed — please login again"))
				}
			} else {
				switch {
				case ctx.Token == "":
					fmt.Fprintf(out, "Status:   %s\n", iostreams.Red("not logged in"))
				case tokenExpired:
					fmt.Fprintf(out, "Status:   %s\n", iostreams.Red("token expired, please login again"))
				default:
					fmt.Fprintf(out, "Status:   %s\n", iostreams.Green("logged in"))
				}
			}

			if !ctx.ExpiresAt.IsZero() {
				label := "Expires:  "
				if tokenExpired {
					label = "Expired:  "
				}
				fmt.Fprintf(out, "%s%s\n", label, ctx.ExpiresAt.Local().Format("2006-01-02 15:04:05"))
			}

			return nil
		},
	}
}
