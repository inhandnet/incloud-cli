package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdDatausageHourly(f *factory.Factory) *cobra.Command {
	opts := &datausageSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "hourly <device-id>",
		Short: "Show hourly data usage",
		Long:  "Display hourly data usage (traffic) for a device. Defaults to today if no time range is specified.",
		Example: `  # Hourly data usage for today (default)
  incloud device datausage hourly 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device datausage hourly 507f1f77bcf86cd799439011 --after 2024-03-01T00:00:00Z --before 2024-03-02T00:00:00Z

  # Filter by traffic type
  incloud device datausage hourly 507f1f77bcf86cd799439011 --type all

  # Table output with selected fields
  incloud device datausage hourly 507f1f77bcf86cd799439011 -o table -f time -f tx -f rx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchDatausageSeries(f, cmd, args[0], "datausage-hourly", opts)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "", "Traffic type: cellular (default), wired, wireless, sim, esim, all, etc.")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00Z)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
