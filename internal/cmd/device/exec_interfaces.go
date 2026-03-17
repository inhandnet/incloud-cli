package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecInterfaces(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "interfaces <device-id>",
		Short:   "List available network interfaces on a device",
		Example: `  incloud device exec interfaces 507f1f77bcf86cd799439011`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return getDiagnosisStatus(f, cmd, args[0], "interfaces")
		},
	}

	return cmd
}
