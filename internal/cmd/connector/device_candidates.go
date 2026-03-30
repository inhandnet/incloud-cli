package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
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

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Search != "" {
				q.Set("nameOrSn", opts.Search)
			}

			body, err := client.Get("/api/v1/connectors/devices/candidates", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or serial number")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")

	return cmd
}
