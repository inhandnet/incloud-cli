package webhook

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Provider string
}

var defaultListFields = []string{"_id", "name", "provider", "webhook", "createdAt"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List webhooks",
		Long:    "List message webhooks with pagination and optional provider filter.",
		Aliases: []string{"ls"},
		Example: `  # List all webhooks
  incloud webhook list

  # Filter by provider
  incloud webhook list --provider wechat

  # Paginate
  incloud webhook list --page 2 --limit 50`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Provider != "" {
				q.Set("provider", opts.Provider)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/message/webhooks", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Filter by provider (supported: wechat)")
	opts.ListFlags.RegisterExpand(cmd)

	return cmd
}
