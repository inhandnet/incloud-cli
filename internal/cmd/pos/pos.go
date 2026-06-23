package pos

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// NewCmdPos creates the `pos` command group for POS Ready management.
func NewCmdPos(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pos",
		Aliases: []string{"posready"},
		Short:   "Manage POS Ready traffic prioritization",
		Long: "Inspect and manage POS Ready traffic prioritization.\n\n" +
			"POS Ready lets you mark which clients' point-of-sale traffic a device should " +
			"prioritize, bypass, or treat normally, and observe which POS applications are matched.\n\n" +
			"To change a single client's level, use 'incloud device client set-pos-ready'.",
	}

	cmd.AddCommand(newCmdClients(f))
	cmd.AddCommand(newCmdForwarded(f))
	cmd.AddCommand(newCmdDeviceHits(f))
	cmd.AddCommand(newCmdMarkedClients(f))
	cmd.AddCommand(newCmdVendorHits(f))
	cmd.AddCommand(newCmdVendorSummary(f))
	cmd.AddCommand(newCmdClientTypes(f))
	cmd.AddCommand(newCmdRules(f))

	return cmd
}
