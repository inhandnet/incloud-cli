package pos

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdDeviceHits(f *factory.Factory) *cobra.Command {
	var (
		activeWithin string
		groupBy      string
	)

	cmd := &cobra.Command{
		Use:   "device-hits <device-id>",
		Short: "Show POS application hits aggregated for a device",
		Long:  "Show POS application traffic hits for a device, aggregated by vendor or by client.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Vendor-grouped hits in the last 24h
  incloud pos device-hits 507f1f77bcf86cd799439011

  # Client-grouped hits over the last 7 days
  incloud pos device-hits 507f1f77bcf86cd799439011 --group-by client --active-within 7d`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if activeWithin != "" {
				q.Set("activeWithin", activeWithin)
			}
			if groupBy != "" {
				q.Set("groupBy", groupBy)
			}

			body, err := client.Get("/api/v1/network/devices/"+args[0]+"/pos-hits", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&activeWithin, "active-within", "", "Time window: 1h, 6h, 24h (default), 7d")
	cmd.Flags().StringVar(&groupBy, "group-by", "", "Aggregation: vendor (default) or client")

	return cmd
}
