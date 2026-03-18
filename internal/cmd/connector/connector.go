package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdConnector(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "connector",
		Aliases: []string{"conn"},
		Short:   "Manage InConnect connector networks",
		Long:    "Create and manage InConnect connector VPN networks, accounts, devices, and endpoints.",
	}

	cmd.AddCommand(newCmdNetwork(f))
	cmd.AddCommand(newCmdAccount(f))
	cmd.AddCommand(newCmdDevice(f))
	cmd.AddCommand(newCmdEndpoint(f))

	return cmd
}
