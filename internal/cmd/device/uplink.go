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

type uplinkOptions struct {
	Fields []string
}

var defaultUplinkFields = []string{"name", "type", "status", "mode", "publicIp", "latency", "loss"}
var defaultUplinkDetailFields = []string{"name", "type", "status", "mode", "publicIp", "latency", "loss", "deviceName"}

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

			reqURL := actx.Host + "/api/v1/devices/" + deviceID + "/uplinks"
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
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
				fields := opts.Fields
				if len(fields) == 0 && f.IO.IsStdoutTTY() {
					fields = defaultUplinkFields
				}
				if err := iostreams.FormatTable(body, f.IO, fields); err != nil {
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

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	cmd.AddCommand(newCmdUplinkGet(f))
	cmd.AddCommand(newCmdUplinkPerf(f))

	return cmd
}
