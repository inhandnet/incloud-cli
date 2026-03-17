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

type uplinkPerfOptions struct {
	Name   string
	After  string
	Before string
	Fields []string
}

var defaultUplinkPerfFields = []string{"time", "throughputUp", "throughputDown", "latency", "jitter", "loss", "signal"}

func newCmdUplinkPerf(f *factory.Factory) *cobra.Command {
	opts := &uplinkPerfOptions{}

	cmd := &cobra.Command{
		Use:   "perf <device-id>",
		Short: "Show uplink performance trend",
		Long:  "Show uplink performance metrics (throughput, latency, jitter, loss) over time for a specific device uplink.",
		Example: `  # Show performance trend for wan1
  incloud device uplink perf 507f1f77bcf86cd799439011 --name wan1

  # Filter by time range
  incloud device uplink perf 507f1f77bcf86cd799439011 --name cellular1 --after 2024-01-01T00:00:00 --before 2024-01-02T00:00:00

  # Table output
  incloud device uplink perf 507f1f77bcf86cd799439011 --name wan1 -o table`,
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

			u, err := url.Parse(actx.Host + "/api/v1/devices/" + deviceID + "/uplinks/perf-trend")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			q.Set("name", opts.Name)
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
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
				flat, err := flattenSeries(body)
				if err != nil {
					return err
				}
				fields := opts.Fields
				if len(fields) == 0 {
					fields = defaultUplinkPerfFields
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

	cmd.Flags().StringVar(&opts.Name, "name", "", "Uplink name (required, e.g. wan1, cellular1)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
