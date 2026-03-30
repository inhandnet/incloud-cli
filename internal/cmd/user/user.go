package user

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdUser(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  "List, create, update, delete, and manage users on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdMe(f))
	cmd.AddCommand(NewCmdIdentity(f))
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdUpdate(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdLock(f))
	cmd.AddCommand(NewCmdUnlock(f))

	return cmd
}
