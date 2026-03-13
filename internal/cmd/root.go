package cmd

import (
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

	cmd.PersistentFlags().StringP("output", "o", "", "Output format: json")
	cmd.PersistentFlags().String("context", "", "Override active context (env: INCLOUD_CONTEXT)")

	// Override INCLOUD_CONTEXT from --context flag
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if ctx, _ := cmd.Flags().GetString("context"); ctx != "" {
			// This makes ActiveContext() pick up the override
			cmd.Flags().Set("context", ctx)
		}
		return nil
	}

	return cmd
}
