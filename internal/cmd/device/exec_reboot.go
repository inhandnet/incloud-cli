package device

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
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
				confirmed, err := confirmPrompt(f, fmt.Sprintf("Reboot device %s?", idsArg))
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(f.IO.ErrOut, "Aborted.")
					return nil
				}
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			ids := strings.Split(idsArg, ",")
			if len(ids) > 1 {
				return bulkInvokeMethod(cmd, f, client, actx.Host, ids, "nezha_reboot", nil)
			}
			return invokeMethod(cmd, f, client, actx.Host, ids[0], "nezha_reboot", 30, nil)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
