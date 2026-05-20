package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdClient(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage touch clients",
		Long:  "Create, list, get, update, delete, and export remote access clients (downstream endpoints).",
	}

	cmd.AddCommand(newCmdClientCreate(f))
	cmd.AddCommand(newCmdClientList(f))
	cmd.AddCommand(newCmdClientGet(f))
	cmd.AddCommand(newCmdClientUpdate(f))
	cmd.AddCommand(newCmdClientDelete(f))
	cmd.AddCommand(newCmdClientExport(f))
	cmd.AddCommand(newCmdClientConnections(f))

	return cmd
}
