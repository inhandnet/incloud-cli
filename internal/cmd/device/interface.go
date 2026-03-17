package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			if opts.Refresh {
				refreshURL := ctx.Host + "/api/v1/devices/" + deviceID + "/interfaces/refresh"
				req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, refreshURL, http.NoBody)
				if err != nil {
					return fmt.Errorf("building refresh request: %w", err)
				}
				resp, err := client.Do(req)
				if err != nil {
					return fmt.Errorf("refresh request failed: %w", err)
				}
				resp.Body.Close()
				if resp.StatusCode == http.StatusRequestTimeout {
					fmt.Fprintln(f.IO.ErrOut, "Warning: refresh timed out (device may be offline), showing cached data")
				} else if resp.StatusCode >= 400 {
					fmt.Fprintln(f.IO.ErrOut, "Warning: refresh failed, showing cached data")
				}
			}

			reqURL := ctx.Host + "/api/v1/devices/" + deviceID + "/interfaces"
			req, err := http.NewRequestWithContext(context.Background(), "GET", reqURL, http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "table":
				flatData, err := flattenInterfaces(body)
				if err != nil {
					return fmt.Errorf("parsing interfaces: %w", err)
				}
				fields := opts.Fields
				if len(fields) == 0 && f.IO.IsStdoutTTY() {
					fields = defaultInterfaceFields
				}
				if err := iostreams.FormatTable(flatData, f.IO, fields); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			}

			return nil
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

	// Wrap as {"result": [...]} so FormatTable's unwrapResult works.
	wrapped := map[string]any{"result": rows}
	return json.Marshal(wrapped)
}
