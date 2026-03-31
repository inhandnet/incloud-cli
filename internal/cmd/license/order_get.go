package license

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OrderGetOptions struct {
	Fields []string
	Expand []string
}

func NewCmdOrderGet(f *factory.Factory) *cobra.Command {
	opts := &OrderGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <order-id>",
		Short: "Get order details",
		Long:  "Get detailed information about a specific order by its ID.",
		Example: `  # View order details
  incloud license order get 64a1b2c3d4e5f6a7b8c9d0e1

  # YAML output
  incloud license order get 64a1b2c3d4e5f6a7b8c9d0e1 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orderID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if len(opts.Fields) > 0 {
				q.Set("fields", strings.Join(opts.Fields, ","))
			}
			if len(opts.Expand) > 0 {
				q.Set("expand", strings.Join(opts.Expand, ","))
			} else {
				q.Set("expand", "org")
			}

			body, err := client.Get("/api/v1/billing/orders/"+orderID, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related resources (supported: creator, org)")

	return cmd
}
