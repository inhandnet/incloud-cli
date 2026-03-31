package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdOrder(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Manage license orders",
		Long:  "List and view license orders on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdOrderList(f))
	cmd.AddCommand(NewCmdOrderGet(f))

	return cmd
}
