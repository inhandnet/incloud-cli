package connector

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type deviceCandidatesOptions struct {
	Search string
	Page   int
	Limit  int
}

func newCmdDeviceCandidates(f *factory.Factory) *cobra.Command {
	opts := &deviceCandidatesOptions{}

	cmd := &cobra.Command{
		Use:   "candidates",
		Short: "List candidate devices that can be added to a connector network",
		Example: `  # List all candidates
  incloud connector device candidates

  # Search by name or serial number
  incloud connector device candidates -q ER805`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Search != "" {
				q.Set("nameOrSn", opts.Search)
			}

			body, err := client.Get("/api/v1/connectors/devices/candidates", q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or serial number")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")

	return cmd
}
