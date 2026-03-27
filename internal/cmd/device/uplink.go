package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type uplinkOptions struct {
	Fields []string
}

var uplinkFormatters = iostreams.ColumnFormatters{
	"latency": iostreams.FormatMicroseconds,
	"jitter":  iostreams.FormatMicroseconds,
}

func NewCmdUplink(f *factory.Factory) *cobra.Command {
	opts := &uplinkOptions{}

	cmd := &cobra.Command{
		Use:   "uplink <device-id>",
		Short: "Show device uplinks",
		Long:  "Show uplink (WAN/Cellular/WiFi) information for a specific device.",
		Example: `  # Show uplinks for a device
  incloud device uplink 507f1f77bcf86cd799439011

  # Table output
  incloud device uplink 507f1f77bcf86cd799439011 -o table

  # Table with selected fields
  incloud device uplink 507f1f77bcf86cd799439011 -o table -f name -f type -f status -f latency`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/uplinks", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithFormatters(uplinkFormatters),
			)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	cmd.AddCommand(newCmdUplinkGet(f))
	cmd.AddCommand(newCmdUplinkPerf(f))

	return cmd
}
