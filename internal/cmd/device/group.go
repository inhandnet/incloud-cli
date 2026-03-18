package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdGroup(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage device groups",
		Long:  "List, create, update, delete, and inspect device groups on the InCloud platform.",
	}

	cmd.AddCommand(newCmdGroupList(f))
	cmd.AddCommand(newCmdGroupGet(f))
	cmd.AddCommand(newCmdGroupCreate(f))
	cmd.AddCommand(newCmdGroupUpdate(f))
	cmd.AddCommand(newCmdGroupDelete(f))
	cmd.AddCommand(newCmdGroupFirmwares(f))

	return cmd
}
