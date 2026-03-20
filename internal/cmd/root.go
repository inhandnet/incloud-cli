package cmd

import (
	"os"

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

	cmd.PersistentFlags().StringP("output", "o", "", "Output format: json, table, yaml")
	cmd.PersistentFlags().String("context", "", "Override active context (env: INCLOUD_CONTEXT)")
	cmd.PersistentFlags().String("sudo", "", "Impersonate a user (env: INCLOUD_SUDO)")
	cmd.PersistentFlags().Lookup("sudo").Hidden = true
	cmd.PersistentFlags().Bool("debug", false, "Enable debug output (env: INCLOUD_DEBUG)")

	// Propagate --context and --sudo flags to env so downstream code picks them up
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Enable debug mode from flag or env var
		if d, _ := cmd.Flags().GetBool("debug"); d {
			debug.Enabled = true
		} else if os.Getenv("INCLOUD_DEBUG") != "" {
			debug.Enabled = true
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
		return nil
	}

	return cmd
}
