package alert

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdRuleDelete(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <rule-id> [<rule-id>...]",
		Short: "Delete alert rules",
		Long:  "Delete one or more alert rules by ID. Multiple IDs triggers bulk delete.",
		Example: `  # Delete a single rule
  incloud alert rule delete 507f1f77bcf86cd799439011

  # Delete multiple rules
  incloud alert rule delete 507f1f77bcf86cd799439011 507f1f77bcf86cd799439012`,
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if len(args) == 1 {
				if _, err := client.Delete("/api/v1/alerts/rules/" + args[0]); err != nil {
					return err
				}
				fmt.Fprintf(f.IO.ErrOut, "Deleted alert rule %s.\n", args[0])
			} else {
				if _, err := client.Post("/api/v1/alerts/rules/bulk-delete", map[string]any{
					"ids": args,
				}); err != nil {
					return err
				}
				fmt.Fprintf(f.IO.ErrOut, "Deleted %d alert rule(s).\n", len(args))
			}

			return nil
		},
	}

	return cmd
}
