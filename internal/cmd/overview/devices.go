package overview

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type DevicesOptions struct {
	Fields []string
}

var defaultDevicesFields = []string{"count.total", "count.online", "count.offline", "count.inactive"}

func NewCmdDevices(f *factory.Factory) *cobra.Command {
	opts := &DevicesOptions{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "Device status distribution",
		Long:  "Show device status distribution including online/offline/inactive counts and product breakdown.",
		Example: `  # Show device status summary
  incloud overview devices

  # Show specific sections
  incloud overview devices -f product -f upgrade

  # JSON output
  incloud overview devices -o json

  # YAML output
  incloud overview devices -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/devices/summary", nil)
			if err != nil {
				return err
			}

			// Unwrap the "result" envelope
			var envelope struct {
				Result json.RawMessage `json:"result"`
			}
			data := body
			if json.Unmarshal(body, &envelope) == nil && envelope.Result != nil {
				data = envelope.Result
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultDevicesFields
			}
			return iostreams.FormatOutput(data, f.IO, output, fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
