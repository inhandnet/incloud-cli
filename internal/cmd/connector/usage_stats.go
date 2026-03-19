package connector

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdUsageStats(f *factory.Factory) *cobra.Command {
	var after, before string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show overall traffic statistics",
		Example: `  # Show traffic statistics for a date range
  incloud connector usage stats --after 2025-01-01 --before 2025-01-31`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", after)
			q.Set("before", before)

			body, err := client.Get("/api/v1/connectors/usage/statistics", q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&before, "before", "", "End date (YYYY-MM-DD, required)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
