package oobm

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdOobm(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oobm",
		Short: "Manage Out-of-Band Management resources",
		Long:  "Create, list, update, delete, connect, and close OOBM resources and serial ports.",
	}

	// Resource commands
	cmd.AddCommand(NewCmdOobmList(f))
	cmd.AddCommand(NewCmdOobmCreate(f))
	cmd.AddCommand(NewCmdOobmUpdate(f))
	cmd.AddCommand(NewCmdOobmDelete(f))
	cmd.AddCommand(NewCmdOobmConnect(f))
	cmd.AddCommand(NewCmdOobmClose(f))
	cmd.AddCommand(NewCmdOobmLogs(f))

	// Serial sub-resource
	cmd.AddCommand(NewCmdOobmSerial(f))

	return cmd
}
