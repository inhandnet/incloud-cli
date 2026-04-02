package connector

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
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
			q.Set("after", cmdutil.ParseTimeFlag(after))
			q.Set("before", cmdutil.ParseTimeFlag(before))
			q.Set("n", strconv.Itoa(n))
			if typ != "" {
				q.Set("type", typ)
			}

			body, err := client.Get("/api/v1/connectors/usage/topk", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (e.g. 2025-01-01, required)")
	cmd.Flags().StringVar(&before, "before", "", "End date (e.g. 2025-01-31, required)")
	cmd.Flags().IntVar(&n, "n", 10, "Number of top results to return")
	cmd.Flags().StringVar(&typ, "type", "", "Filter by type: DEVICE or ACCOUNT (default: DEVICE)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
