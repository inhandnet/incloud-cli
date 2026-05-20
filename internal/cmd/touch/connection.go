package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdConnection(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage touch connections",
		Long:  "Create and disconnect remote access connections to touch clients.",
	}

	cmd.AddCommand(newCmdConnectionCreate(f))
	cmd.AddCommand(newCmdConnectionDisconnect(f))

	return cmd
}
