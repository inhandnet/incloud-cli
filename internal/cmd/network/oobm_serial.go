package network

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdOobmSerial(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serial",
		Short: "Manage OOBM serial port configurations",
	}

	cmd.AddCommand(NewCmdOobmSerialList(f))
	cmd.AddCommand(NewCmdOobmSerialCreate(f))
	cmd.AddCommand(NewCmdOobmSerialUpdate(f))
	cmd.AddCommand(NewCmdOobmSerialDelete(f))
	cmd.AddCommand(NewCmdOobmSerialConnect(f))
	cmd.AddCommand(NewCmdOobmSerialClose(f))

	return cmd
}
