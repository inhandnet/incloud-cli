package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
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

			// Fetch current user and org from API when logged in
			if ctx.EffectiveToken() != "" && !tokenExpired {
				client, err := f.APIClient()
				if err == nil {
					q := url.Values{}
					q.Set("fields", "username,email")
					q.Set("expand", "org")
					body, err := client.Get("/api/v1/users/me", q)
					if err == nil {
						var resp struct {
							Result struct {
								Username string `json:"username"`
								Email    string `json:"email"`
								Org      struct {
									Name string `json:"name"`
									ID   string `json:"_id"`
								} `json:"org"`
							} `json:"result"`
						}
						if json.Unmarshal(body, &resp) == nil {
							me := resp.Result
							if me.Username != "" {
								fmt.Fprintf(out, "Account:  %s", me.Username)
								if me.Email != "" {
									fmt.Fprintf(out, " (%s)", me.Email)
								}
								fmt.Fprintln(out)
							}
							if me.Org.Name != "" {
								fmt.Fprintf(out, "Org:      %s (%s)\n", me.Org.Name, me.Org.ID)
							}
						}
					}
				}
			}

			return nil
		},
	}
}
