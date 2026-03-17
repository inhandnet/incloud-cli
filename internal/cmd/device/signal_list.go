package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type signalListOptions struct {
	After  string
	Before string
	Fields []string
}

var defaultSignalFields = []string{"time", "type", "rsrp", "rsrq", "sinr", "networkType", "carrier", "band"}

func newCmdSignalList(f *factory.Factory) *cobra.Command {
	opts := &signalListOptions{}

	cmd := &cobra.Command{
		Use:   "list <device-id>",
		Short: "Show signal quality over time",
		Long:  "Display signal quality metrics (RSRP, RSRQ, SINR, etc.) for a device over time.",
		Example: `  # Show signal data for a device
  incloud device signal list 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device signal list 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00 --before 2024-01-02T00:00:00

  # Table output with selected fields
  incloud device signal list 507f1f77bcf86cd799439011 -o table -f time -f rsrp -f sinr`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID + "/signal")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
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
				flat, err := flattenSignalSeries(body)
				if err != nil {
					return err
				}
				fields := opts.Fields
				if len(fields) == 0 && f.IO.IsStdoutTTY() {
					fields = defaultSignalFields
				}
				if err := iostreams.FormatTable(flat, f.IO, fields); err != nil {
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

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}

// flattenSignalSeries uses the shared flattenSeries with includeType=true
// since signal series have a "type" field per series (e.g. "4G", "5G").
func flattenSignalSeries(body []byte) ([]byte, error) {
	return flattenSeriesWithType(body)
}
