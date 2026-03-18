package connector

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type accountEventsOptions struct {
	After  string
	Before string
}

func newCmdAccountEvents(f *factory.Factory) *cobra.Command {
	opts := &accountEventsOptions{}

	cmd := &cobra.Command{
		Use:   "events <network-id> <account-id>",
		Short: "Show account online/offline events",
		Example: `  # Show events in a time range
  incloud connector account events <network-id> <account-id> --after 2025-01-01T00:00:00 --before 2025-01-31T23:59:59`,
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

			body, err := client.Get("/api/v1/connectors/"+networkID+"/accounts/"+accountID+"/online-events", q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, required)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, required)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
