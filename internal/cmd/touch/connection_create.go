package touch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConnectionCreate(f *factory.Factory) *cobra.Command {
	var (
		deviceID  string
		username  string
		clientIDs []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create touch connections",
		Long:  "Create remote access connections to one or more touch clients on a device.",
		Example: `  # Connect to clients on a device
  incloud touch connection create --device-id 507f1f77bcf86cd799439011 --username 653b1ff2a84e171614d88695 --client-ids id1,id2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"deviceId":  deviceID,
				"username":  username,
				"clientIds": clientIDs,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/touch/connections", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Connections created.\n")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&deviceID, "device-id", "", "Target device ID (required)")
	cmd.Flags().StringVar(&username, "username", "", "User token ID (required)")
	cmd.Flags().StringSliceVar(&clientIDs, "client-ids", nil, "Client IDs to connect to (required, comma-separated)")
	_ = cmd.MarkFlagRequired("device-id")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("client-ids")

	return cmd
}
