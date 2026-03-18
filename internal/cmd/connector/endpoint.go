package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdEndpoint(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "endpoint",
		Aliases: []string{"ep"},
		Short:   "Manage endpoints in connector networks",
	}

	cmd.AddCommand(newCmdEndpointList(f))
	cmd.AddCommand(newCmdEndpointCreate(f))
	cmd.AddCommand(newCmdEndpointUpdate(f))
	cmd.AddCommand(newCmdEndpointDelete(f))

	return cmd
}
