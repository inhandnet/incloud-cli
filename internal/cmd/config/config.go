package config

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(NewCmdCurrentContext(f))
	cmd.AddCommand(NewCmdListContexts(f))
	cmd.AddCommand(NewCmdUseContext(f))
	cmd.AddCommand(NewCmdSetContext(f))
	cmd.AddCommand(NewCmdDeleteContext(f))

	return cmd
}
