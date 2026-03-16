package cmd

import (
	"os"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdRoot(f *factory.Factory, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "incloud",
		Short:         "InCloud Platform CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	cmd.PersistentFlags().StringP("output", "o", "", "Output format: json, table, yaml")
	cmd.PersistentFlags().String("context", "", "Override active context (env: INCLOUD_CONTEXT)")

	// Propagate --context flag to env so config.ActiveContext() picks it up
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if ctx, _ := cmd.Flags().GetString("context"); ctx != "" {
			os.Setenv("INCLOUD_CONTEXT", ctx)
		}
		return nil
	}

	return cmd
}
