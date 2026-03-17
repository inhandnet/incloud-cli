package device

import (
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

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("name", opts.Name)
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/uplinks/perf-trend", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 {
				fields = defaultUplinkPerfFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fields, iostreams.WithTransform(iostreams.FlattenSeries))
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Uplink name (required, e.g. wan1, cellular1)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
