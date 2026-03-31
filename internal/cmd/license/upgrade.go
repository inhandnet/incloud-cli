package license

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func NewCmdUpgrade(f *factory.Factory) *cobra.Command {
	var (
		deviceIDs []string
		to        string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade license type on devices",
		Long: `Upgrade licenses on devices from one type to a higher tier.

All selected devices must have the same current license type, and none can have
an expired license. Remaining days are recalculated based on the price ratio
between the current and target license types.`,
		Example: `  # Upgrade licenses on two devices to professional type
  incloud license upgrade --device-id DEV_ID1,DEV_ID2 --to professional

  # Upgrade with confirmation skipped
  incloud license upgrade --device-id DEV_ID1 --to enterprise --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(deviceIDs) > 1000 {
				return fmt.Errorf("maximum 1000 devices per upgrade operation, got %d", len(deviceIDs))
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"deviceIds":   deviceIDs,
				"licenseType": to,
			}

			previewResp, err := client.Post("/api/v1/billing/licenses/pre-upgrade", body)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if err := iostreams.FormatOutput(previewResp, f.IO, output); err != nil {
				return err
			}

			if !yes {
				confirmed, err := ui.Confirm(f, fmt.Sprintf("Upgrade %d device(s) to license type %q?", len(deviceIDs), to))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			_, err = client.Post("/api/v1/billing/licenses/upgrade", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Upgraded %d device(s) to license type %q.\n", len(deviceIDs), to)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&deviceIDs, "device-id", nil, "Device IDs to upgrade (required, comma-separated)")
	_ = cmd.MarkFlagRequired("device-id")
	cmd.Flags().StringVar(&to, "to", "", "Target license type slug (required)")
	_ = cmd.MarkFlagRequired("to")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
