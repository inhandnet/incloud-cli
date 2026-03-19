package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecPing(f *factory.Factory) *cobra.Command {
	var (
		host       string
		iface      string
		source     string
		packetSize int
		pingCount  int
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "ping <device-id>",
		Short: "Run ping diagnostic on a device",
		Long: `Run ping from a remote device and stream results in real time.

The command starts a ping task on the device, subscribes to the result stream,
and prints each line as it arrives — similar to running ping locally.
Press Ctrl+C to cancel.`,
		Example: `  # Ping a host from the device (uses default interface "any", 4 packets)
  incloud device exec ping 507f1f77bcf86cd799439011 --host 8.8.8.8

  # With specific interface and count
  incloud device exec ping 507f1f77bcf86cd799439011 --host 8.8.8.8 --interface eth0 --count 10

  # Output raw JSON response instead of streaming
  incloud device exec ping 507f1f77bcf86cd799439011 --host 8.8.8.8 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params := map[string]any{
				"host":       host,
				"interface":  iface,
				"packetSize": packetSize,
				"pingCount":  pingCount,
				"source":     source,
			}

			if jsonOutput {
				return runDiagnosis(f, cmd, args[0], "ping", params)
			}
			return runDiagnosisStream(f, args[0], "ping", params)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Target host to ping (required)")
	cmd.Flags().StringVar(&iface, "interface", "any",
		"Network interface to use (use 'incloud device exec interfaces <device-id>' to list available interfaces)")
	cmd.Flags().StringVar(&source, "source", "", "Source address (only effective when interface is \"any\")")
	cmd.Flags().IntVar(&packetSize, "packet-size", 64, "Packet size in bytes (1-65535)")
	cmd.Flags().IntVar(&pingCount, "count", 4, "Number of ping packets (1-1000)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output raw JSON response instead of streaming results")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}
