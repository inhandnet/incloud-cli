package device

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type uplinkPerfOptions struct {
	Name   string
	After  string
	Before string
	Fields []string
}

func newCmdUplinkPerf(f *factory.Factory) *cobra.Command {
	opts := &uplinkPerfOptions{}

	cmd := &cobra.Command{
		Use:   "perf <device-id>",
		Short: "Show uplink performance trend",
		Long: `Show uplink performance metrics (throughput, latency, jitter, loss) over time for a specific device uplink.

The loss field is a fraction (0.05 = 5%), not a percentage.

-o json / -o yaml / --jq output is columnar: field names appear in "columns" and
the samples in "values", positionally aligned, one array per row. A
"latencyStatus" / "jitterStatus" column is appended only when at least one
sample is not a measurement:
  "timeout"   the probe timed out for that sample; the value is null.

Table output is flattened into row objects with different, shorter field
names.`,
		Example: `  # Show performance trend for wan1
  incloud device uplink perf 507f1f77bcf86cd799439011 --name wan1

  # Filter by time range
  incloud device uplink perf 507f1f77bcf86cd799439011 --name cellular1 --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z

  # JSON output
  incloud device uplink perf 507f1f77bcf86cd799439011 --name wan1 -o json`,
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
				q.Set("after", cmdutil.ParseTimeFlag(opts.After))
			}
			if opts.Before != "" {
				q.Set("before", cmdutil.ParseTimeFlag(opts.Before))
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/uplinks/perf-trend", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if !cmd.Flags().Changed("output") {
				output = "table"
			}
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(iostreams.FlattenSeries),
				iostreams.WithFormatters(iostreams.ColumnFormatters{
					"throughputUp":   iostreams.FormatBitRate,
					"throughputDown": iostreams.FormatBitRate,
					"latency":        iostreams.FormatMicroseconds,
					"jitter":         iostreams.FormatMicroseconds,
				}),
				iostreams.WithDeclaredUnits("device uplink perf"),
			)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Uplink name (required, e.g. wan1, cellular1)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
