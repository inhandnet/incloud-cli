package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecPing(f *factory.Factory) *cobra.Command {
	var (
		host       string
		iface      string
		packetSize int
		pingCount  int
	)

	cmd := &cobra.Command{
		Use:   "ping <device-id>",
		Short: "Run ping diagnostic on a device",
		Example: `  # Ping a host from the device
  incloud device exec ping 507f1f77bcf86cd799439011 --host 8.8.8.8

  # With specific interface and count
  incloud device exec ping 507f1f77bcf86cd799439011 --host 8.8.8.8 --interface eth0 --count 10`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnosis(f, cmd, args[0], "ping", map[string]interface{}{
				"host":       host,
				"interface":  iface,
				"packetSize": packetSize,
				"pingCount":  pingCount,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Target host to ping (required)")
	cmd.Flags().StringVar(&iface, "interface", "", "Network interface to use")
	cmd.Flags().IntVar(&packetSize, "packet-size", 0, "Packet size in bytes")
	cmd.Flags().IntVar(&pingCount, "count", 0, "Number of ping packets")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}
