package device

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecSpeedtestConfig(f *factory.Factory) *cobra.Command {
	var iface string

	cmd := &cobra.Command{
		Use:   "speedtest-config <device-id>",
		Short: "Get speed test configuration (available interfaces and server nodes)",
		Example: `  # List available interfaces and server nodes
  incloud device exec speedtest-config 507f1f77bcf86cd799439011

  # Refresh server nodes for a specific interface
  incloud device exec speedtest-config 507f1f77bcf86cd799439011 --interface eth0`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := url.Values{}
			if iface != "" {
				q.Set("interface", iface)
			}
			return getDiagnosisStatus(f, cmd, args[0], "speedtest/config", q)
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "", "Network interface to scan server nodes for")

	return cmd
}
