package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdConfigAbort(f *factory.Factory) *cobra.Command {
	var (
		module string
		yes    bool
	)

	cmd := &cobra.Command{
		Use:   "abort <device-id>",
		Short: "Abort pending configuration delivery",
		Long:  "Abort pending configuration delivery for a device, clearing the delta between desired and reported state so the cloud accepts the device's current configuration.",
		Example: `  # Abort pending config delivery (with confirmation)
  incloud device config abort 507f1f77bcf86cd799439011

  # Skip confirmation
  incloud device config abort 507f1f77bcf86cd799439011 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Abort pending config delivery for device %s?", deviceID))
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

			path := "/api/v1/devices/" + deviceID + "/pending/config"
			if module != "" {
				path += "?module=" + module
			}

			body, err := client.Delete(path)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				fmt.Fprintf(f.IO.ErrOut, "Pending configuration delivery aborted for device %s.\n", deviceID)
				return nil
			}
			return iostreams.FormatOutput(body, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
