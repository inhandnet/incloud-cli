package connector

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdUsageTopK(f *factory.Factory) *cobra.Command {
	var after, before, typ string
	var n int

	cmd := &cobra.Command{
		Use:   "topk",
		Short: "Show top-K traffic consumption ranking",
		Example: `  # Show top 10 devices by traffic
  incloud connector usage topk --after 2025-01-01 --before 2025-01-31

  # Show top 5 accounts by traffic
  incloud connector usage topk --after 2025-01-01 --before 2025-01-31 --type ACCOUNT --n 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if typ != "" && typ != "DEVICE" && typ != "ACCOUNT" {
				return fmt.Errorf("invalid type %q: must be DEVICE or ACCOUNT", typ)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", after)
			q.Set("before", before)
			q.Set("n", strconv.Itoa(n))
			if typ != "" {
				q.Set("type", typ)
			}

			body, err := client.Get("/api/v1/connectors/usage/topk", q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (YYYY-MM-DD, required)")
	cmd.Flags().StringVar(&before, "before", "", "End date (YYYY-MM-DD, required)")
	cmd.Flags().IntVar(&n, "n", 10, "Number of top results to return")
	cmd.Flags().StringVar(&typ, "type", "", "Filter by type: DEVICE or ACCOUNT (default: DEVICE)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
