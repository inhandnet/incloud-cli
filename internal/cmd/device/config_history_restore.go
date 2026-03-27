package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/inhandnet/incloud-cli/internal/ui"
)

func newCmdConfigHistoryRestore(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "restore <device-id> <snapshot-id>",
		Short: "Restore configuration from a snapshot",
		Long:  "Restore a device's configuration from a history snapshot. This replaces the current device-level configuration with the snapshot's merged config.",
		Example: `  # Restore with confirmation
  incloud device config snapshots restore 507f1f77bcf86cd799439011 69ba26b4ed93070787cea168

  # Skip confirmation
  incloud device config snapshots restore 507f1f77bcf86cd799439011 69ba26b4ed93070787cea168 --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]
			snapshotID := args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if !yes {
				// Fetch snapshot metadata for a meaningful confirmation prompt
				prompt := fmt.Sprintf("Restore config from snapshot %s for device %s?", snapshotID, deviceID)

				snapshotPath := fmt.Sprintf("/api/v1/devices/%s/config/history/%s", deviceID, snapshotID)
				if snapBody, err := client.Get(snapshotPath, nil); err == nil {
					var snap struct {
						Result struct {
							CreatedAt string `json:"createdAt"`
							Trigger   string `json:"trigger"`
						} `json:"result"`
					}
					if json.Unmarshal(snapBody, &snap) == nil && snap.Result.CreatedAt != "" {
						prompt = fmt.Sprintf("Restore config from snapshot %s (trigger: %s, created: %s) for device %s?",
							snapshotID, snap.Result.Trigger, snap.Result.CreatedAt, deviceID)
					}
				}

				confirmed, err := ui.Confirm(f, prompt)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			}

			path := fmt.Sprintf("/api/v1/devices/%s/config/history/%s/apply", deviceID, snapshotID)
			body, err := client.Post(path, nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				fmt.Fprintf(f.IO.ErrOut, "Configuration restored from snapshot %s.\n", snapshotID)
				return nil
			}
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
