package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClientDatausageDaily(f *factory.Factory) *cobra.Command {
	opts := &clientSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "datausage-daily <client-id>",
		Short: "Client daily data usage",
		Long:  "Display daily data usage (tx/rx) for a client in a given month.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Current month
  incloud device client datausage-daily 507f1f77bcf86cd799439011

  # Specific month
  incloud device client datausage-daily 507f1f77bcf86cd799439011 --month 2026-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchClientSeries(f, cmd, args[0], "datausage-daily", opts, defaultClientDatausageFields, clientDatausageFormatters)
		},
	}

	cmd.Flags().StringVar(&opts.Month, "month", "", "Month in YYYY-MM format (e.g. 2026-03)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
