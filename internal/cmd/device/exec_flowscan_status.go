package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecFlowscanStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "flowscan-status <device-id>",
		Short:   "Get flow scan status for a device",
		Example: `  incloud device exec flowscan-status 507f1f77bcf86cd799439011`,
		Args:    cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "device id", "incloud device list -q %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getDiagnosisStatus(f, cmd, args[0], "flowscan")
		},
	}

	return cmd
}
