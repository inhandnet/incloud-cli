package device

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type clientOnlineEventsOptions struct {
	Page   int
	Limit  int
	After  string
	Before string
	Fields []string
}

func newCmdClientOnlineEvents(f *factory.Factory) *cobra.Command {
	opts := &clientOnlineEventsOptions{}

	cmd := &cobra.Command{
		Use:   "online-events <client-id>",
		Short: "Client connect/disconnect events",
		Long:  "List online/offline events (connect and disconnect history) for a client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}

			body, err := client.Get("/api/v1/network/clients/"+args[0]+"/online-events-list", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
