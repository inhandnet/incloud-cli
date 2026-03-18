package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdNetwork(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "network",
		Aliases: []string{"net"},
		Short:   "Manage connector networks",
	}

	cmd.AddCommand(newCmdNetworkList(f))
	cmd.AddCommand(newCmdNetworkGet(f))
	cmd.AddCommand(newCmdNetworkCreate(f))
	cmd.AddCommand(newCmdNetworkUpdate(f))
	cmd.AddCommand(newCmdNetworkDelete(f))
	cmd.AddCommand(newCmdNetworkStats(f))

	return cmd
}
