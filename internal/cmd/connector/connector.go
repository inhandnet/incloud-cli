package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdConnector(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "connector",
		Aliases: []string{"conn"},
		Short:   "Manage InCloud Manager connector networks",
		Long:    "Create and manage InCloud Manager connector VPN networks, accounts, devices, and endpoints.",
	}

	cmd.AddCommand(newCmdNetwork(f))
	cmd.AddCommand(newCmdAccount(f))
	cmd.AddCommand(newCmdDevice(f))
	cmd.AddCommand(newCmdEndpoint(f))
	cmd.AddCommand(newCmdUsage(f))

	return cmd
}
