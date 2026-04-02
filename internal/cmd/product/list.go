package product

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Name   string
	Type   string
	Status string
}

var defaultListFields = []string{"_id", "name", "productType", "status", "deprecated"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List products",
		Long:    "List products on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List products with default pagination
  incloud product list

  # Paginate
  incloud product list --page 2 --limit 50

  # Filter by name (LIKE search)
  incloud product list --name IR615

  # Filter by product type
  incloud product list --type router

  # Filter by status
  incloud product list --status PUBLISHED

  # Table output with selected fields
  incloud product list -o table -f name -f productType -f status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.Type != "" {
				q.Set("productType", opts.Type)
			}
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/products", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by product type")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (INDEVELOPMENT|PUBLISHED)")

	return cmd
}
