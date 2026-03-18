package network

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdNetwork(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage network services",
		Long:  "Manage network services including OOBM, VPN, connectors, and more.",
	}

	cmd.AddCommand(NewCmdOobm(f))

	return cmd
}
