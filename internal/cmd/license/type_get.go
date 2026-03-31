package license

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TypeGetOptions struct {
	Fields []string
	Expand []string
}

func NewCmdTypeGet(f *factory.Factory) *cobra.Command {
	opts := &TypeGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <slug>",
		Short: "Get license type details",
		Long:  "Get detailed information about a specific license type by its slug.",
		Example: `  # View license type details
  incloud license type get professional

  # YAML output
  incloud license type get professional -o yaml

  # Only specific fields
  incloud license type get professional -f name -f premiumServices -f upgrades`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

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
			}

			body, err := client.Get("/api/v1/billing/license-types/"+slug, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related resources (supported: prices)")

	return cmd
}
