package model

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type listOptions struct {
	cmdutil.ListFlags
	Name          string
	Tags          []string
	ProductModels []string
}

func newCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List AI models",
		Long:    "List AI models in the current tenant.",
		Aliases: []string{"ls"},
		Example: `  # List all models
  incloud model list

  # Search by name (fuzzy match)
  incloud model list --name detect

  # Filter by tags
  incloud model list --tags edge,ai

  # Filter by product model
  incloud model list --product-models IR315-xxx

  # Show specific fields
  incloud model list --fields _id,name,tags`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if len(opts.Tags) > 0 {
				for _, tag := range opts.Tags {
					q.Add("tags", tag)
				}
			}
			if len(opts.ProductModels) > 0 {
				for _, pm := range opts.ProductModels {
					q.Add("productModels", pm)
				}
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/models", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (fuzzy match)")
	cmd.Flags().StringSliceVar(&opts.Tags, "tags", nil, "Filter by tags")
	cmd.Flags().StringSliceVar(&opts.ProductModels, "product-models", nil, "Filter by product models (format: product-pn)")

	return cmd
}
