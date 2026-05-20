package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdEdge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edge",
		Short: "Manage edge device properties",
		Long:  "View and manage edge-specific device properties including environment variables, project versions, and CLI configuration.",
	}

	cmd.AddCommand(newCmdEdgeList(f))
	cmd.AddCommand(newCmdEdgeGet(f))
	cmd.AddCommand(newCmdEdgeUpdate(f))
	cmd.AddCommand(newCmdEdgePin(f))
	cmd.AddCommand(newCmdEdgeCliConfig(f))

	return cmd
}
