package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecFlowscan(f *factory.Factory) *cobra.Command {
	var (
		iface     string
		duration  int
		flowType  string
		srcFilter []string
	)

	cmd := &cobra.Command{
		Use:   "flowscan <device-id>",
		Short: "Start flow scan (traffic analysis) on a device",
		Example: `  # Start flow scan
  incloud device exec flowscan 507f1f77bcf86cd799439011

  # With duration and type filter
  incloud device exec flowscan 507f1f77bcf86cd799439011 --duration 7200 --type DOMAIN`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnosis(f, cmd, args[0], "flowscan", map[string]interface{}{
				"interface": iface,
				"duration":  duration,
				"type":      flowType,
				"srcFilter": srcFilter,
			})
		},
	}

	cmd.Flags().StringVar(&iface, "interface", "", "Network interface to use")
	cmd.Flags().IntVar(&duration, "duration", 0, "Scan duration in seconds (3600-259200)")
	cmd.Flags().StringVar(&flowType, "type", "", "Flow type: ALL or DOMAIN")
	cmd.Flags().StringSliceVar(&srcFilter, "src-filter", nil, "Source IP filters (comma-separated)")

	return cmd
}
