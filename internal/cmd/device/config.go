package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Device configuration management",
		Long:  "View, update, and manage device configurations including layered config, history snapshots, and error diagnostics.",
	}

	cmd.AddCommand(newCmdConfigGet(f))
	cmd.AddCommand(newCmdConfigError(f))
	cmd.AddCommand(newCmdConfigHistory(f))
	cmd.AddCommand(newCmdConfigAbort(f))
	cmd.AddCommand(newCmdConfigUpdate(f))
	cmd.AddCommand(newCmdConfigCopy(f))

	return cmd
}
