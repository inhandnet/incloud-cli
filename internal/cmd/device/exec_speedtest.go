package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecSpeedtest(f *factory.Factory) *cobra.Command {
	var (
		iface      string
		serverNode string
	)

	cmd := &cobra.Command{
		Use:   "speedtest <device-id>",
		Short: "Run speed test on a device",
		Example: `  # Run speed test
  incloud device exec speedtest 507f1f77bcf86cd799439011

  # With specific interface and server
  incloud device exec speedtest 507f1f77bcf86cd799439011 --interface eth0 --server-node node1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnosis(f, cmd, args[0], "speedtest", map[string]interface{}{
				"interface":  iface,
				"serverNode": serverNode,
			})
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "", "Network interface to use")
	cmd.Flags().StringVar(&serverNode, "server-node", "", "Speed test server node ID")

	return cmd
}
