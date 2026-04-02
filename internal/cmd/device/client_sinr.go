package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

var defaultClientSINRFields = []string{"time", "sinr"}

func newCmdClientSINR(f *factory.Factory) *cobra.Command {
	opts := &clientSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "sinr <client-id>",
		Short: "Client signal-to-interference ratio (SINR)",
		Long:  "Display SINR (Signal to Interference plus Noise Ratio) time-series data for a wireless client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchClientSeries(f, cmd, args[0], "sinr", opts, defaultClientSINRFields)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
