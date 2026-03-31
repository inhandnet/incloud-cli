package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdLicense(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "license",
		Short: "Manage licenses",
		Long:  "Query, assign, and manage licenses on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdHistory(f))
	cmd.AddCommand(NewCmdType(f))
	cmd.AddCommand(NewCmdOrder(f))
	cmd.AddCommand(NewCmdAttach(f))
	cmd.AddCommand(NewCmdDetach(f))
	cmd.AddCommand(NewCmdUpgrade(f))
	cmd.AddCommand(NewCmdTransfer(f))
	cmd.AddCommand(NewCmdAlignExpiry(f))

	return cmd
}
