package license

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdDetach(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "detach <license-id>",
		Short: "Detach a license from its device",
		Long:  "Detach a license from the device it is currently attached to.",
		Example: `  # Detach a license (will prompt for confirmation)
  incloud license detach 64a1b2c3d4e5f6a7b8c9d0e1

  # Skip confirmation
  incloud license detach 64a1b2c3d4e5f6a7b8c9d0e1 --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			licenseID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			licenseResp, err := client.Get("/api/v1/billing/licenses/"+licenseID, nil)
			if err != nil {
				return err
			}

			lic, err := parseLicenseState(licenseResp)
			if err != nil {
				return err
			}

			if lic.DeviceID == "" {
				return fmt.Errorf("license %s is not attached to any device", licenseID)
			}
			if lic.Status != "activated" && lic.Status != "to_be_expired" {
				return fmt.Errorf("cannot detach: license %s has status %q (must be activated or to_be_expired)", licenseID, lic.Status)
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Detach license %s from device %s?", licenseID, lic.DeviceID))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Delete("/api/v1/billing/licenses/" + licenseID + "/device")
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "License %s detached from device %s.\n", licenseID, lic.DeviceID)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
