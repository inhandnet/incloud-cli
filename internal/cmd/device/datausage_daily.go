package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdDatausageDaily(f *factory.Factory) *cobra.Command {
	opts := &datausageSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "daily <device-id>",
		Short: "Show daily data usage",
		Long:  "Display daily data usage (traffic) for a device. Defaults to the current month if no month is specified.",
		Example: `  # Daily data usage for current month
  incloud device datausage daily 507f1f77bcf86cd799439011

  # Specify month
  incloud device datausage daily 507f1f77bcf86cd799439011 --month 2024-03

  # Filter by traffic type
  incloud device datausage daily 507f1f77bcf86cd799439011 --type all`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchDatausageSeries(f, cmd, args[0], "datausage-daily", opts)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "", "Traffic type: cellular (default), wired, wireless, sim, esim, all, etc.")
	cmd.Flags().StringVar(&opts.Month, "month", "", "Month to query (YYYY-MM, e.g. 2024-03)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
