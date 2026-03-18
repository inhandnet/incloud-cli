package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdTransfer(f *factory.Factory) *cobra.Command {
	var (
		org string
		yes bool
	)

	cmd := &cobra.Command{
		Use:   "transfer <device-id>",
		Short: "Transfer a device to another organization",
		Long: `Transfer a device to another organization.

This is a destructive operation: the device record is deleted from the
source organization and recreated in the target organization.`,
		Example: `  # Transfer device to another organization
  incloud device transfer 507f1f77bcf86cd799439011 --org 60b76cb7ee4e6d5d842429da

  # Skip confirmation
  incloud device transfer 507f1f77bcf86cd799439011 --org 60b76cb7ee4e6d5d842429da -y`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Transfer device %s to organization %s? This cannot be undone.", deviceID, org))
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

			body := map[string]interface{}{
				"deviceIds": []string{deviceID},
				"to":        org,
			}

			_, err = client.Put("/api/v1/devices/transfer", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Device %s transferred to organization %s.\n", deviceID, org)
			return nil
		},
	}

	cmd.Flags().StringVar(&org, "org", "", "Target organization ID (required)")
	_ = cmd.MarkFlagRequired("org")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
