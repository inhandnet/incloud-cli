package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdHistory(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <license-id>",
		Short: "View license operation history",
		Long:  "View the operation log for a specific license, including attach, detach, upgrade, align, and expire events.",
		Example: `  # View operation history
  incloud license history YFE5QYOTHKHBMSX

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

			body, err := client.Get("/api/v1/billing/licenses/"+licenseID+"/history", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
