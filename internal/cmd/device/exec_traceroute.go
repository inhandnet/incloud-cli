package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecTraceroute(f *factory.Factory) *cobra.Command {
	var (
		host  string
		iface string
	)

	cmd := &cobra.Command{
		Use:   "traceroute <device-id>",
		Short: "Run traceroute diagnostic on a device",
		Example: `  # Traceroute to a host from the device
  incloud device exec traceroute 507f1f77bcf86cd799439011 --host 8.8.8.8`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnosis(f, cmd, args[0], "traceroute", map[string]interface{}{
				"host":      host,
				"interface": iface,
			})
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Target host (required)")
	cmd.Flags().StringVar(&iface, "interface", "", "Network interface to use")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}
