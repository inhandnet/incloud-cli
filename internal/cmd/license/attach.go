package license

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdAttach(f *factory.Factory) *cobra.Command {
	var (
		device string
		yes    bool
	)

	cmd := &cobra.Command{
		Use:   "attach <license-id>",
		Short: "Attach a license to a device",
		Long: `Attach a license to a device.

If the device already has an active license of the same type, the new license
duration will be added to the existing one (overlay). This operation is irreversible.`,
		Example: `  # Attach a license to a device
  incloud license attach 64a1b2c3d4e5f6a7b8c9d0e1 --device 507f1f77bcf86cd799439011

  # Skip confirmation for overlay
  incloud license attach 64a1b2c3d4e5f6a7b8c9d0e1 --device 507f1f77bcf86cd799439011 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			licenseID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Get("/api/v1/devices/"+device, nil)
			if err != nil {
				return err
			}
			devLic, err := parseDeviceLicense(resp)
			if err != nil {
				return err
			}

			if devLic.ID != "" && (devLic.Status == "activated" || devLic.Status == "to_be_expired") {
				if !yes {
					if !ui.IsTTY(f) {
						return fmt.Errorf(
							"device %s already has an active license (%s). Attaching will overlay (add duration). This is irreversible. Use --yes to confirm",
							device, devLic.ID,
						)
					}
					confirmed, err := ui.Confirm(f, fmt.Sprintf(
						"Device %s already has an active license (%s). Attaching will overlay (add duration). This is irreversible. Continue?",
						device, devLic.ID,
					))
					if err != nil {
						return err
					}
					if !confirmed {
						return nil
					}
				}
			}

			_, err = client.Put("/api/v1/billing/licenses/"+licenseID+"/device", map[string]any{
				"deviceId": device,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "License %s attached to device %s.\n", licenseID, device)
			return nil
		},
	}

	cmd.Flags().StringVar(&device, "device", "", "Target device ID (required)")
	_ = cmd.MarkFlagRequired("device")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
