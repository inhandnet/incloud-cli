package license

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
	Expand []string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <license-id>",
		Short: "Get license details",
		Long:  "Get detailed information about a specific license by its ID.",
		Example: `  # View license details
  incloud license get 64a1b2c3d4e5f6a7b8c9d0e1

  # View only status and expiry
  incloud license get 64a1b2c3d4e5f6a7b8c9d0e1 -f status -f expiresAt

  # YAML output
  incloud license get 64a1b2c3d4e5f6a7b8c9d0e1 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			licenseID := args[0]

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
				q.Set("expand", "org,device")
			}

			body, err := client.Get("/api/v1/billing/licenses/"+licenseID, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related resources (supported: device, org)")

	return cmd
}
