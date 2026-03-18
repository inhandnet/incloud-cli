package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultClientDatausageFields = []string{"time", "tx", "rx", "total"}

var clientDatausageFormatters = iostreams.WithFormatters(iostreams.ColumnFormatters{
	"tx":    iostreams.FormatBytes,
	"rx":    iostreams.FormatBytes,
	"total": iostreams.FormatBytes,
})

func newCmdClientDatausageHourly(f *factory.Factory) *cobra.Command {
	opts := &clientSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "datausage-hourly <client-id>",
		Short: "Client hourly data usage",
		Long:  "Display hourly data usage (tx/rx) time-series for a client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchClientSeries(f, cmd, args[0], "datausage-hourly", opts, defaultClientDatausageFields, clientDatausageFormatters)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
