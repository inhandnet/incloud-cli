package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

var defaultClientThroughputFields = []string{"time", "throughputUp", "throughputDown"}

func newCmdClientThroughput(f *factory.Factory) *cobra.Command {
	opts := &clientSeriesOptions{}

	cmd := &cobra.Command{
		Use:   "throughput <client-id>",
		Short: "Client throughput over time",
		Long:  "Display upload and download throughput time-series data for a client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fetchClientSeries(f, cmd, args[0], "throughput", opts, defaultClientThroughputFields)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, required)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, required)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
