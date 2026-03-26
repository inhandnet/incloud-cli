package tunnel

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTunnel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Manage remote access tunnels",
		Long:  "Open and close remote access tunnels (Web UI / CLI) for devices.",
	}

	cmd.AddCommand(NewCmdTunnelOpenWeb(f))
	cmd.AddCommand(NewCmdTunnelOpenCli(f))
	cmd.AddCommand(NewCmdTunnelForward(f))
	cmd.AddCommand(NewCmdTunnelClose(f))
	cmd.AddCommand(NewCmdTunnelLogs(f))

	return cmd
}
