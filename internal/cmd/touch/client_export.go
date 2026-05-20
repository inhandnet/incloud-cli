package touch

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClientExport(f *factory.Factory) *cobra.Command {
	var (
		deviceID string
		status   string
		name     string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export touch clients",
		Long:  "Export touch clients to a file (CSV/Excel).",
		Example: `  # Export all clients to a file
  incloud touch client export --out clients.csv

  # Export clients for a specific device
  incloud touch client export --device-id 507f1f77bcf86cd799439011 --out clients.csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			path := "/api/v1/touch/clients/export"
			sep := "?"
			if deviceID != "" {
				path += sep + "deviceId=" + deviceID
				sep = "&"
			}
			if status != "" {
				path += sep + "touchConnectionStatus=" + status
				sep = "&"
			}
			if name != "" {
				path += sep + "name=" + name
			}

			if err := client.Download(path, output); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Exported to %s.\n", output)
			return nil
		},
	}

	cmd.Flags().StringVar(&deviceID, "device-id", "", "Filter by device ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by connection status")
	cmd.Flags().StringVar(&name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&output, "out", "clients.csv", "Output file path")

	return cmd
}
