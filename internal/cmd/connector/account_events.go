package connector

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
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
  incloud connector account events <network-id> <account-id> --after 2025-01-01T00:00:00Z --before 2025-01-31T23:59:59Z`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID, accountID := args[0], args[1]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("after", cmdutil.ParseTimeFlag(opts.After))
			q.Set("before", cmdutil.ParseTimeFlag(opts.Before))

			body, err := client.Get("/api/v1/connectors/"+networkID+"/accounts/"+accountID+"/online-events", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z)")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
