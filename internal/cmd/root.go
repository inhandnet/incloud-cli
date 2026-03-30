package cmd

import (
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/build"
	"github.com/inhandnet/incloud-cli/internal/debug"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "incloud",
		Short:         "InCloud Platform CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       build.Version,
	}

	cmd.PersistentFlags().StringP("output", "o", "", "Output format: json, table, yaml (default: table for TTY, json otherwise)")
	cmd.PersistentFlags().String("jq", "", `Filter JSON output using a jq expression (implies -o json)`)
	cmd.PersistentFlags().String("context", "", "Override active context (env: INCLOUD_CONTEXT)")
	cmd.PersistentFlags().String("sudo", "", "Impersonate a user (env: INCLOUD_SUDO)")
	cmd.PersistentFlags().Lookup("sudo").Hidden = true
	cmd.PersistentFlags().String("tenant", "", "Switch organization context by ID (env: INCLOUD_TENANT)")
	cmd.PersistentFlags().Bool("debug", false, "Enable debug output (env: INCLOUD_DEBUG)")

	// Propagate flags to env and set output default based on TTY
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Enable debug mode from flag or env var
		if d, _ := cmd.Flags().GetBool("debug"); d {
			debug.Enabled = true
		} else if os.Getenv("INCLOUD_DEBUG") != "" {
			debug.Enabled = true
		}

		// Default output format: table for TTY, json for pipes/redirects
		outputExplicit := cmd.Flags().Changed("output")
		if !outputExplicit {
			if f.IO.IsStdoutTTY() {
				_ = cmd.Flags().Set("output", "table")
			} else {
				_ = cmd.Flags().Set("output", "json")
			}
		}
		if outputExplicit {
			cmd.Flags().Lookup("output").Annotations = map[string][]string{"explicit": {"true"}}
		}

		// --jq implies JSON output mode
		if jqExpr, _ := cmd.Flags().GetString("jq"); jqExpr != "" {
			f.IO.JQExpr = jqExpr
			if !outputExplicit {
				_ = cmd.Flags().Set("output", "json")
			}
		}

		if ctx, _ := cmd.Flags().GetString("context"); ctx != "" {
			if err := os.Setenv("INCLOUD_CONTEXT", ctx); err != nil {
				return err
			}
		}
		if sudo, _ := cmd.Flags().GetString("sudo"); sudo != "" {
			if err := os.Setenv("INCLOUD_SUDO", sudo); err != nil {
				return err
			}
		}
		if tenant, _ := cmd.Flags().GetString("tenant"); tenant != "" {
			if err := os.Setenv("INCLOUD_TENANT", tenant); err != nil {
				return err
			}
		}
		return nil
	}

	return cmd
}

// SetupSuperAdminFlags unhides super-admin-only flags (e.g. --sudo)
// when the current user is a super admin.
// The result is cached to a file per context with a 1-hour TTL to avoid
// repeated API calls across CLI invocations.
func SetupSuperAdminFlags(rootCmd *cobra.Command, f *factory.Factory) {
	if isSuperAdmin(f) {
		if fl := rootCmd.PersistentFlags().Lookup("sudo"); fl != nil {
			fl.Hidden = false
		}
	}
}

const superAdminCacheTTL = 1 * time.Hour

// isSuperAdmin checks the config cache first, falls back to API, then persists.
func isSuperAdmin(f *factory.Factory) bool {
	cfg, err := f.Config()
	if err != nil {
		return false
	}
	ctx, err := cfg.ActiveContext()
	if err != nil {
		return false
	}

	// Use cached value if present and valid
	if ctx.SuperAdmin != nil &&
		!ctx.SuperAdminAt.IsZero() &&
		!ctx.SuperAdminAt.After(time.Now()) &&
		time.Since(ctx.SuperAdminAt) < superAdminCacheTTL {
		debug.Log("admin check: using cached value %v", *ctx.SuperAdmin)
		return *ctx.SuperAdmin
	}

	result := checkSuperAdmin(f)

	// Persist to config
	ctx.SuperAdmin = &result
	ctx.SuperAdminAt = time.Now()
	_ = f.SaveConfig()

	return result
}

func checkSuperAdmin(f *factory.Factory) bool {
	client, err := f.APIClient()
	if err != nil {
		debug.Log("admin check: failed to get API client: %v", err)
		return false
	}

	q := url.Values{}
	q.Set("fields", "roles")
	body, err := client.Get("/api/v1/users/me", q)
	if err != nil {
		debug.Log("admin check: API request failed: %v", err)
		return false
	}

	var resp struct {
		Result struct {
			Roles []struct {
				Name        string `json:"name"`
				BuiltInRole bool   `json:"builtInRole"`
			} `json:"roles"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		debug.Log("admin check: failed to parse response: %v", err)
		return false
	}

	for _, role := range resp.Result.Roles {
		if role.Name == "root" && role.BuiltInRole {
			return true
		}
	}
	return false
}
