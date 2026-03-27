package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type InterfaceOptions struct {
	Refresh bool
	Fields  []string
}

var defaultInterfaceFields = []string{"name", "type", "state", "subnet", "publicIp"}

// interfaceTypes lists the keys we extract from the API result, in display order.
var interfaceTypes = []string{"cellular", "wan", "lan", "wifiSta"}

func NewCmdInterface(f *factory.Factory) *cobra.Command {
	opts := &InterfaceOptions{}

	cmd := &cobra.Command{
		Use:   "interface <device-id>",
		Short: "Show device network interfaces",
		Long:  "Show network interface information for a specific device.",
		Example: `  # Show interfaces as JSON
  incloud device interface 507f1f77bcf86cd799439011

  # Real-time refresh (device must be online)
  incloud device interface 507f1f77bcf86cd799439011 --refresh

  # Table output
  incloud device interface 507f1f77bcf86cd799439011 -o table

  # Table with selected fields
  incloud device interface 507f1f77bcf86cd799439011 -o table -f name -f type -f state -f gateway`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if opts.Refresh {
				_, err := client.Post("/api/v1/devices/"+deviceID+"/interfaces/refresh", nil)
				if err != nil {
					fmt.Fprintln(f.IO.ErrOut, "Warning: refresh failed, showing cached data")
				}
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/interfaces", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultInterfaceFields
			}
			return iostreams.FormatOutput(body, f.IO, output, iostreams.WithTransform(flattenInterfaces))
		},
	}

	cmd.Flags().BoolVar(&opts.Refresh, "refresh", false, "Trigger real-time collection (device must be online)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}

// flattenInterfaces extracts interface arrays from the API response and merges
// them into a single JSON array, adding a "type" field to each entry.
func flattenInterfaces(data []byte) ([]byte, error) {
	var envelope struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}

	var rows []map[string]any
	for _, ifType := range interfaceTypes {
		raw, ok := envelope.Result[ifType]
		if !ok {
			continue
		}
		var ifaces []map[string]any
		if err := json.Unmarshal(raw, &ifaces); err != nil {
			continue
		}
		for _, iface := range ifaces {
			iface["type"] = ifType
			rows = append(rows, iface)
		}
	}

	// Wrap as {"result": [...]} so FormatTable extracts the array.
	wrapped := map[string]any{"result": rows}
	return json.Marshal(wrapped)
}
