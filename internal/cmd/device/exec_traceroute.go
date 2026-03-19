package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecTraceroute(f *factory.Factory) *cobra.Command {
	var (
		host       string
		iface      string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "traceroute <device-id>",
		Short: "Run traceroute diagnostic on a device",
		Long: `Run traceroute from a remote device and stream results in real time.

The command starts a traceroute task on the device, subscribes to the result stream,
and prints each line as it arrives. Press Ctrl+C to cancel.`,
		Example: `  # Traceroute to a host from the device
  incloud device exec traceroute 507f1f77bcf86cd799439011 --host 8.8.8.8

  # Output raw JSON response instead of streaming
  incloud device exec traceroute 507f1f77bcf86cd799439011 --host 8.8.8.8 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			params := map[string]any{
				"host":      host,
				"interface": iface,
			}

			if jsonOutput {
				return runDiagnosis(f, cmd, args[0], "traceroute", params)
			}
			return runDiagnosisStream(f, args[0], "traceroute", params)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Target host (required)")
	cmd.Flags().StringVar(&iface, "interface", "any",
		"Network interface to use (use 'incloud device exec interfaces <device-id>' to list available interfaces)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output raw JSON response instead of streaming results")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}
