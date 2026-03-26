package connector

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type accountLogsOptions struct {
	After  string
	Before string
	Page   int
	Limit  int
}

func newCmdAccountLogs(f *factory.Factory) *cobra.Command {
	opts := &accountLogsOptions{}

	cmd := &cobra.Command{
		Use:   "logs <network-id> <account-id>",
		Short: "Show account connection logs",
		Example: `  # Show connection logs
  incloud connector account logs <network-id> <account-id> --after 2025-01-01T00:00:00Z --before 2025-01-31T23:59:59Z`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID, accountID := args[0], args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", opts.After)
			q.Set("before", opts.Before)
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))

			body, err := client.Get("/api/v1/connectors/"+networkID+"/accounts/"+accountID+"/online-logs", q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, required)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, required)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
