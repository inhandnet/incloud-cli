package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdSdwan(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sdwan",
		Short: "Manage SD-WAN networks and devices",
		Long:  "Create and manage SD-WAN (AutoVPN) networks, tunnels, connections, and devices.",
	}

	cmd.AddCommand(newCmdNetwork(f))
	cmd.AddCommand(newCmdDevices(f))
	cmd.AddCommand(newCmdCandidates(f))
	cmd.AddCommand(newCmdVerifySubnets(f))
	cmd.AddCommand(newCmdDeviceSubnets(f))

	return cmd
}
