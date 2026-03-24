package device

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type perfOptions struct {
	After   string
	Before  string
	Refresh bool
	Fields  []string
}

var perfFormatters = iostreams.ColumnFormatters{
	"cpu.usage":      iostreams.FormatPercent,
	"memory.free":    iostreams.FormatBytes,
	"memory.total":   iostreams.FormatBytes,
	"disk.free":      iostreams.FormatBytes,
	"disk.total":     iostreams.FormatBytes,
	"microSD.free":   iostreams.FormatBytes,
	"microSD.total":  iostreams.FormatBytes,
	"msata.free":     iostreams.FormatBytes,
	"msata.total":    iostreams.FormatBytes,
}

func NewCmdPerf(f *factory.Factory) *cobra.Command {
	opts := &perfOptions{}

	cmd := &cobra.Command{
		Use:   "perf <device-id>",
		Short: "Show device performance (CPU, memory, disk)",
		Long: `Show device performance metrics.

By default, displays the current performance snapshot (CPU usage, memory, disk).
With --after/--before, displays historical performance time series.
With --refresh, triggers a real-time collection from the device (online only).`,
		Example: `  # Current performance snapshot
  incloud device perf 507f1f77bcf86cd799439011

  # Real-time collection (device must be online)
  incloud device perf 507f1f77bcf86cd799439011 --refresh

  # Historical time series
  incloud device perf 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00 --before 2024-01-02T00:00:00

  # Select specific fields
  incloud device perf 507f1f77bcf86cd799439011 -f cpu.usage -f memory.free -f memory.total

  # JSON output
  incloud device perf 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if opts.After != "" || opts.Before != "" {
				return runPerfSeries(f, cmd, deviceID, opts)
			}
			return runPerfCurrent(f, cmd, deviceID, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().BoolVar(&opts.Refresh, "refresh", false, "Trigger real-time collection (device must be online)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}

func runPerfCurrent(f *factory.Factory, cmd *cobra.Command, deviceID string, opts *perfOptions) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// Trigger real-time collection first, then always GET cached data
	if opts.Refresh {
		_, err := client.Post("/api/v1/devices/"+deviceID+"/performances/refresh", nil)
		if err != nil {
			fmt.Fprintln(f.IO.ErrOut, "Warning: refresh failed (device may be offline), showing cached data")
		}
	}

	body, err := client.Get("/api/v1/devices/"+deviceID+"/performances", nil)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, f.IO, output, opts.Fields,
		iostreams.WithFormatters(perfFormatters),
	)
}

func runPerfSeries(f *factory.Factory, cmd *cobra.Command, deviceID string, opts *perfOptions) error {
	if opts.After == "" || opts.Before == "" {
		return fmt.Errorf("both --after and --before are required for historical data")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("after", opts.After)
	q.Set("before", opts.Before)

	body, err := client.Get("/api/v1/devices/"+deviceID+"/performance", q)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	return iostreams.FormatOutput(body, f.IO, output, opts.Fields,
		iostreams.WithTransform(iostreams.FlattenSeries),
		iostreams.WithFormatters(perfFormatters),
	)
}
