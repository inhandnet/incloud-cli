package connector

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type deviceUpdateOptions struct {
	Subnet string
}

func newCmdDeviceUpdate(f *factory.Factory) *cobra.Command {
	opts := &deviceUpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <network-id> <device-id>",
		Short: "Update a device in a connector network",
		Example: `  # Update device subnet
  incloud connector device update <network-id> <device-id> --subnet 10.32.35.0/24`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID, deviceID := args[0], args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})
			if cmd.Flags().Changed("subnet") {
				body["subnet"] = opts.Subnet
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify --subnet")
			}

			respBody, err := client.Put("/api/v1/connectors/"+networkID+"/devices/"+deviceID, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			writeUpdated(f, "Connector device", respBody)
			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Subnet, "subnet", "", "Device subnet (e.g. 10.32.35.0/24)")

	return cmd
}
