package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigHistoryGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <device-id> <snapshot-id>",
		Short: "Get a configuration history snapshot",
		Long:  "Get the full details of a configuration history snapshot, including the merged configuration at that point in time.",
		Example: `  # View a snapshot
  incloud device config snapshots get 507f1f77bcf86cd799439011 69ba26b4ed93070787cea168

  # YAML output
  incloud device config snapshots get 507f1f77bcf86cd799439011 69ba26b4ed93070787cea168 -o yaml`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]
			snapshotID := args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/api/v1/devices/%s/config/history/%s", deviceID, snapshotID)
			body, err := client.Get(path, nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	return cmd
}
