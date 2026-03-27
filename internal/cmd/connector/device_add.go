package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type deviceAddOptions struct {
	DeviceID string
	Subnet   string
}

func newCmdDeviceAdd(f *factory.Factory) *cobra.Command {
	opts := &deviceAddOptions{}

	cmd := &cobra.Command{
		Use:   "add <network-id>",
		Short: "Add a device to a connector network",
		Example: `  # Add device to network
  incloud connector device add <network-id> --device-id <device-id>

  # With subnet
  incloud connector device add <network-id> --device-id <device-id> --subnet 10.32.34.0/24`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"deviceId": opts.DeviceID,
			}
			if opts.Subnet != "" {
				body["subnet"] = opts.Subnet
			}

			respBody, err := client.Post("/api/v1/connectors/"+networkID+"/devices", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			writeCreated(f, "Connector device", respBody)
			return formatOutput(cmd, f.IO, respBody)
		},
	}

	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Device ID to add (required; use 'incloud connector device candidates' to find IDs)")
	cmd.Flags().StringVar(&opts.Subnet, "subnet", "", "Device subnet (e.g. 10.32.34.0/24)")

	_ = cmd.MarkFlagRequired("device-id")

	return cmd
}
