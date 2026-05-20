package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdGroupRegistry(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage device group container registries",
		Long:  "Get and update container registry configurations for a device group.",
	}

	cmd.AddCommand(newCmdGroupRegistryGet(f))
	cmd.AddCommand(newCmdGroupRegistryUpdate(f))

	return cmd
}
