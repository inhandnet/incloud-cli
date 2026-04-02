package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var validIntervals = map[string]bool{
	"hourly":  true,
	"daily":   true,
	"monthly": true,
}

func NewCmdDatausage(f *factory.Factory) *cobra.Command {
	opts := &datausageSeriesOptions{}
	interval := "daily"

	cmd := &cobra.Command{
		Use:     "datausage [device-id]",
		Aliases: []string{"du"},
		Short:   "Device data usage statistics",
		Long:    "View device data usage (traffic) statistics at hourly, daily, or monthly granularity.",
		Example: `  # Daily data usage (default interval)
  incloud device datausage 507f1f77bcf86cd799439011

  # Hourly data usage
  incloud device datausage 507f1f77bcf86cd799439011 --interval hourly

  # Monthly data usage for a specific year
  incloud device datausage 507f1f77bcf86cd799439011 --interval monthly --year 2024`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if !validIntervals[interval] {
				return fmt.Errorf("invalid interval %q: must be hourly, daily, or monthly", interval)
			}
			return fetchDatausageSeries(f, cmd, args[0], "datausage-"+interval, opts)
		},
	}

	cmd.Flags().StringVar(&interval, "interval", "daily", "Granularity: hourly, daily, monthly")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Traffic type: cellular (default), wired, wireless, sim, esim, all, etc.")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z)")
	cmd.Flags().StringVar(&opts.Month, "month", "", "Month to query (YYYY-MM, e.g. 2024-03)")
	cmd.Flags().StringVar(&opts.Year, "year", "", "Year to query (YYYY, e.g. 2024)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	cmd.AddCommand(newCmdDatausageHourly(f))
	cmd.AddCommand(newCmdDatausageDaily(f))
	cmd.AddCommand(newCmdDatausageMonthly(f))
	cmd.AddCommand(newCmdDatausageList(f))

	return cmd
}

// datausageSeriesOptions holds flags shared by hourly/daily/monthly subcommands.
// Each subcommand only registers the flags it needs (e.g. daily registers --month
// but not --after/--before).
type datausageSeriesOptions struct {
	Type   string
	After  string
	Before string
	Month  string
	Year   string
	Fields []string
}

// fetchDatausageSeries is the shared RunE for hourly/daily/monthly commands.
// endpoint is the URL path suffix (e.g. "datausage-hourly").
func fetchDatausageSeries(f *factory.Factory, cmd *cobra.Command, deviceID, endpoint string, opts *datausageSeriesOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body, err := client.Get("/api/v1/devices/"+deviceID+"/"+endpoint, url.Values{
		"type":   {opts.Type},
		"after":  {cmdutil.ParseTimeFlag(opts.After)},
		"before": {cmdutil.ParseTimeFlag(opts.Before)},
		"month":  {opts.Month},
		"year":   {opts.Year},
	})
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, f.IO, output,
		iostreams.WithTransform(iostreams.FlattenSeries),
		iostreams.WithFormatters(iostreams.ColumnFormatters{
			"tx":    iostreams.FormatBytes,
			"rx":    iostreams.FormatBytes,
			"total": iostreams.FormatBytes,
		}),
	)
}
