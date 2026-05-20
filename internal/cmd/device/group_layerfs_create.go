package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupLayerfsCreate(f *factory.Factory) *cobra.Command {
	var (
		name        string
		deviceID    string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create <group-id>",
		Short: "Create a filesystem snapshot",
		Long:  "Create a filesystem snapshot (layerfs) by capturing the current filesystem state from a specified edge device.",
		Example: `  # Create a layerfs from a device
  incloud device group layerfs create 507f1f77bcf86cd799439011 --name my-snapshot --device-id 653b1ff2a84e171614d88695

  # With description
  incloud device group layerfs create 507f1f77bcf86cd799439011 --name my-snapshot --device-id 653b1ff2a84e171614d88695 --description "Base image v1"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"name":     name,
				"deviceId": deviceID,
			}
			if description != "" {
				reqBody["description"] = description
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/devicegroups/"+args[0]+"/layerfs", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Layerfs snapshot created in group %s.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Snapshot name (required, 1-64 chars)")
	cmd.Flags().StringVar(&deviceID, "device-id", "", "Source device ID to capture from (required)")
	cmd.Flags().StringVar(&description, "description", "", "Snapshot description (max 128 chars)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("device-id")

	return cmd
}
