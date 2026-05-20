package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdGroupLayerfs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "layerfs",
		Short: "Manage device group filesystem snapshots",
		Long:  "Create, list, update, and delete filesystem snapshots (layerfs) captured from edge devices within a device group.",
	}

	cmd.AddCommand(newCmdGroupLayerfsCreate(f))
	cmd.AddCommand(newCmdGroupLayerfsList(f))
	cmd.AddCommand(newCmdGroupLayerfsUpdate(f))
	cmd.AddCommand(newCmdGroupLayerfsDelete(f))

	return cmd
}
