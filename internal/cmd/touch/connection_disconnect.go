package touch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConnectionDisconnect(f *factory.Factory) *cobra.Command {
	var (
		deviceID  string
		clientIDs []string
	)

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect touch connections",
		Long:  "Disconnect remote access connections for specified clients on a device.",
		Example: `  # Disconnect clients
  incloud touch connection disconnect --device-id 507f1f77bcf86cd799439011 --client-ids id1,id2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"deviceId":  deviceID,
				"clientIds": clientIDs,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Post("/api/v1/touch/connections/disconnect", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Connections disconnected.\n")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&deviceID, "device-id", "", "Target device ID (required)")
	cmd.Flags().StringSliceVar(&clientIDs, "client-ids", nil, "Client IDs to disconnect (required, comma-separated)")
	_ = cmd.MarkFlagRequired("device-id")
	_ = cmd.MarkFlagRequired("client-ids")

	return cmd
}
