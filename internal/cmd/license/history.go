package license

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdHistory(f *factory.Factory) *cobra.Command {
	var expand []string

	cmd := &cobra.Command{
		Use:   "history <license-id>",
		Short: "View license operation history",
		Long:  "View the operation log for a specific license, including attach, detach, upgrade, align, and expire events.",
		Example: `  # View operation history
  incloud license history YFE5QYOTHKHBMSX

  # Include device details in each history entry
  incloud license history YFE5QYOTHKHBMSX --expand device

  # YAML output
  incloud license history YFE5QYOTHKHBMSX -o yaml

  # Filter with jq
  incloud license history YFE5QYOTHKHBMSX --jq '.[] | {type, createdAt}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			licenseID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if len(expand) > 0 {
				q.Set("expand", strings.Join(expand, ","))
			}

			body, err := client.Get("/api/v1/billing/licenses/"+licenseID+"/history", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVar(&expand, "expand", nil, "Expand related resources (supported: device)")

	return cmd
}
