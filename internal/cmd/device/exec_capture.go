package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecCapture(f *factory.Factory) *cobra.Command {
	var (
		iface         string
		captureTime   int
		source        string
		expertOptions string
	)

	cmd := &cobra.Command{
		Use:   "capture <device-id>",
		Short: "Start packet capture (tcpdump) on a device",
		Example: `  # Capture on a specific interface
  incloud device exec capture 507f1f77bcf86cd799439011 --interface eth0

  # With duration and source filter
  incloud device exec capture 507f1f77bcf86cd799439011 --interface eth0 --duration 60 --source 192.168.1.1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnosis(f, cmd, args[0], "capture", map[string]interface{}{
				"interface":     iface,
				"captureTime":   captureTime,
				"source":        source,
				"expertOptions": expertOptions,
			})
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "", "Network interface (required)")
	cmd.Flags().IntVar(&captureTime, "duration", 0, "Capture duration in seconds")
	cmd.Flags().StringVar(&source, "source", "", "Source IP filter")
	cmd.Flags().StringVar(&expertOptions, "expert-options", "", "Advanced tcpdump options")
	_ = cmd.MarkFlagRequired("interface")

	return cmd
}
