package pos

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdRules(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage POS custom rules",
		Long:  "Get, set, and list per-device POS custom rules (add/mask entries layered on the global rule file).",
	}

	cmd.AddCommand(newCmdRulesGet(f))
	cmd.AddCommand(newCmdRulesSet(f))
	cmd.AddCommand(newCmdRulesList(f))

	return cmd
}
