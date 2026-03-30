package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type RuleListOptions struct {
	cmdutil.ListOpts
}

var defaultRuleListFields = []string{"_id", "groupIds", "rules", "notify.channels", "createdAt"}

func NewCmdRuleList(f *factory.Factory) *cobra.Command {
	opts := &RuleListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List alert rules",
		Long:    "List alert rules with pagination.",
		Aliases: []string{"ls"},
		Example: `  # List alert rules
  incloud alert rule list

  # Paginate
  incloud alert rule list --page 2 --limit 50

  # Table output
  incloud alert rule list -o table

  # Table with selected fields
  incloud alert rule list -o table -f _id -f groupIds -f rules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultRuleListFields)

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/alerts/rules", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmdutil.RegisterListFlags(cmd, &opts.ListOpts)

	return cmd
}
