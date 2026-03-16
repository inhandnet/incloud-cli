package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdRule(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rule",
		Short: "Manage alert rules",
		Long:  "Create, list, update, and delete alert rules on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdRuleList(f))
	cmd.AddCommand(NewCmdRuleGet(f))
	cmd.AddCommand(NewCmdRuleCreate(f))
	cmd.AddCommand(NewCmdRuleUpdate(f))
	cmd.AddCommand(NewCmdRuleDelete(f))

	return cmd
}
