package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdExecReboot(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "reboot <id>[,<id2>,...]",
		Short: "Reboot a device",
		Example: `  # Reboot a single device
  incloud device exec reboot 507f1f77bcf86cd799439011

  # Reboot multiple devices
  incloud device exec reboot 507f1f77bcf86cd799439011,653b1ff2a84e171614d88695

  # Skip confirmation
  incloud device exec reboot 507f1f77bcf86cd799439011 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idsArg := args[0]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Reboot device %s?", idsArg))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			ids := strings.Split(idsArg, ",")
			if len(ids) > 1 {
				return bulkInvokeMethod(cmd, f, client, ids, "nezha_reboot", nil)
			}
			return invokeMethod(cmd, f, client, ids[0], "nezha_reboot", 30, nil)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
