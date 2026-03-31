package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TypeListOptions struct {
	cmdutil.ListFlags
	Product string
	Status  string
}

var defaultTypeListFields = []string{"slug", "name", "products", "status", "publishedAt"}

func NewCmdTypeList(f *factory.Factory) *cobra.Command {
	opts := &TypeListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List license types",
		Long:    "List available license types with optional filtering by product and status.",
		Aliases: []string{"ls"},
		Example: `  # List all published license types
  incloud license type list

  # Filter by product
  incloud license type list --product IR915

  # Include unpublished types
  incloud license type list --status unpublished

  # YAML output
  incloud license type list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultTypeListFields)
			if opts.Product != "" {
				q.Set("product", opts.Product)
			}
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/billing/license-types", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Product, "product", "", "Filter by product name")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (published/unpublished, default: published)")
	opts.ListFlags.RegisterExpand(cmd, "prices")

	return cmd
}
