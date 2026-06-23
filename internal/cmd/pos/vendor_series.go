package pos

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// fetchVendorSeries is the shared RunE body for the two device+client POS
// vendor time-series endpoints (pos-vendor-hits, pos-vendor-summary), both of
// which return a FluxResult flattened for table output.
func fetchVendorSeries(f *factory.Factory, cmd *cobra.Command, deviceID, clientID, endpoint, after, before, interval string) error {
	if after == "" || before == "" {
		return fmt.Errorf("both --after and --before are required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("after", cmdutil.ParseTimeFlag(after))
	q.Set("before", cmdutil.ParseTimeFlag(before))
	if interval != "" {
		q.Set("interval", interval)
	}

	body, err := client.Get("/api/v1/network/devices/"+deviceID+"/clients/"+clientID+"/"+endpoint, q)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	if !cmd.Flags().Changed("output") {
		output = "table"
	}
	return iostreams.FormatOutput(body, f.IO, output, iostreams.WithTransform(iostreams.FlattenSeries))
}

func newCmdVendorHits(f *factory.Factory) *cobra.Command {
	var (
		after    string
		before   string
		interval string
	)

	cmd := &cobra.Command{
		Use:   "vendor-hits <device-id> <client-id>",
		Short: "POS vendor hit time series for a client",
		Long:  "Show the POS vendor hit time series for a client on a device, bucketed by interval.",
		Args:  cobra.ExactArgs(2),
		Example: `  # 5-minute buckets over a day
  incloud pos vendor-hits 507f1f77bcf86cd799439011 69b8c537e7f8d2c1e5fffdbc \
    --after 2026-03-17T00:00:00Z --before 2026-03-18T00:00:00Z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchVendorSeries(f, cmd, args[0], args[1], "pos-vendor-hits", after, before, interval)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start time (RFC3339 or YYYY-MM-DD), required")
	cmd.Flags().StringVar(&before, "before", "", "End time (RFC3339 or YYYY-MM-DD), required")
	cmd.Flags().StringVar(&interval, "interval", "", "Bucket interval (default 5m)")

	return cmd
}

func newCmdVendorSummary(f *factory.Factory) *cobra.Command {
	var (
		after  string
		before string
	)

	cmd := &cobra.Command{
		Use:   "vendor-summary <device-id> <client-id>",
		Short: "POS vendor hit summary for a client",
		Long:  "Show the aggregated POS vendor hit summary for a client on a device over a time range.",
		Args:  cobra.ExactArgs(2),
		Example: `  # Summary over a day
  incloud pos vendor-summary 507f1f77bcf86cd799439011 69b8c537e7f8d2c1e5fffdbc \
    --after 2026-03-17T00:00:00Z --before 2026-03-18T00:00:00Z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchVendorSeries(f, cmd, args[0], args[1], "pos-vendor-summary", after, before, "")
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start time (RFC3339 or YYYY-MM-DD), required")
	cmd.Flags().StringVar(&before, "before", "", "End time (RFC3339 or YYYY-MM-DD), required")

	return cmd
}
