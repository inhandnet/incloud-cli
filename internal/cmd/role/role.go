package role

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdRole(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role",
		Short: "Manage roles",
		Long:  "List and manage roles on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))

	return cmd
}
