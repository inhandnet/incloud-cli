package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdRuleGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <rule-id>",
		Short: "Get alert rule details",
		Long:  "Get detailed information about a specific alert rule by its ID.",
		Example: `  # Get rule details
  incloud alert rule get 507f1f77bcf86cd799439011

  # Table output
  incloud alert rule get 507f1f77bcf86cd799439011 -o table

  # YAML output
  incloud alert rule get 507f1f77bcf86cd799439011 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ruleID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/alerts/rules/"+ruleID, nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
