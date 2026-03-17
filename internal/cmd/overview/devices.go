package overview

import (
	"encoding/json"
	"fmt"

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
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}
			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			apiURL := actx.Host + "/api/v1/devices/summary"

			body, err := doGet(client, apiURL)
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
			switch output {
			case "table":
				fields := opts.Fields
				if len(fields) == 0 && f.IO.IsStdoutTTY() {
					fields = defaultDevicesFields
				}
				if err := iostreams.FormatTable(data, f.IO, fields); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(data)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(data) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(data, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(data))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
