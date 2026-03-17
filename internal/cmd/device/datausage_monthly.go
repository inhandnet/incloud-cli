package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdDatausageMonthly(f *factory.Factory) *cobra.Command {
	opts := &datausageSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "monthly <device-id>",
		Short: "Show monthly data usage",
		Long:  "Display monthly data usage (traffic) for a device. Defaults to the current year if no year is specified.",
		Example: `  # Monthly data usage for current year
  incloud device datausage monthly 507f1f77bcf86cd799439011

  # Specify year
  incloud device datausage monthly 507f1f77bcf86cd799439011 --year 2024

  # Filter by traffic type
  incloud device datausage monthly 507f1f77bcf86cd799439011 --type all`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchDatausageSeries(f, cmd, args[0], "datausage-monthly", opts)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "", "Traffic type: cellular (default), wired, wireless, sim, esim, all, etc.")
	cmd.Flags().StringVar(&opts.Year, "year", "", "Year to query (YYYY, e.g. 2024)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
