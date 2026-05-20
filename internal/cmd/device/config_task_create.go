package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigTaskCreate(f *factory.Factory) *cobra.Command {
	var (
		product   string
		deviceIDs []string
		groupID   string
		config    string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a CLI configuration task",
		Long:  "Create a task to push CLI configuration commands to edge devices. Target by device IDs or group ID.",
		Example: `  # Push config to specific devices
  incloud device config task create --product IR615 --device-ids id1,id2 --config "interface cellular 1\n ip nat inside"

  # Push config to a device group
  incloud device config task create --product IR615 --group-id 507f1f77bcf86cd799439011 --config "interface cellular 1\n ip nat inside"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"product": product,
				"config":  config,
			}
			if len(deviceIDs) > 0 {
				reqBody["deviceIds"] = deviceIDs
			}
			if groupID != "" {
				reqBody["groupId"] = groupID
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/live/cli-configs", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "CLI configuration task created.\n")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&product, "product", "", "Product type (required)")
	cmd.Flags().StringSliceVar(&deviceIDs, "device-ids", nil, "Target device IDs (comma-separated, max 100)")
	cmd.Flags().StringVar(&groupID, "group-id", "", "Target device group ID")
	cmd.Flags().StringVar(&config, "config", "", "CLI configuration commands to push (required)")
	_ = cmd.MarkFlagRequired("product")
	_ = cmd.MarkFlagRequired("config")

	return cmd
}
