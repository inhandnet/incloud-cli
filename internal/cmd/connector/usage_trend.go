package connector

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdUsageTrend(f *factory.Factory) *cobra.Command {
	var after, before string

	cmd := &cobra.Command{
		Use:   "trend",
		Short: "Show daily traffic trend",
		Example: `  # Show daily traffic trend for a date range
  incloud connector usage trend --after 2025-01-01 --before 2025-01-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", cmdutil.ParseTimeFlag(after))
			q.Set("before", cmdutil.ParseTimeFlag(before))

			body, err := client.Get("/api/v1/connectors/usage/tendency", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if !cmd.Flags().Changed("output") {
				output = "table"
			}
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(iostreams.FlattenSeries))
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (e.g. 2025-01-01, required)")
	cmd.Flags().StringVar(&before, "before", "", "End date (e.g. 2025-01-31, required)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
