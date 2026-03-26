package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdDevice(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"dev"},
		Short:   "Manage devices",
		Long:    "List, create, update, delete, and manage devices on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdUpdate(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdExport(f))
	cmd.AddCommand(NewCmdImport(f))
	cmd.AddCommand(NewCmdImportStatus(f))
	cmd.AddCommand(NewCmdAssign(f))
	cmd.AddCommand(NewCmdUnassign(f))
	cmd.AddCommand(NewCmdTransfer(f))
	cmd.AddCommand(NewCmdSignal(f))
	cmd.AddCommand(NewCmdAntenna(f))
	cmd.AddCommand(NewCmdPerf(f))
	cmd.AddCommand(NewCmdOnline(f))
	cmd.AddCommand(NewCmdLog(f))
	cmd.AddCommand(NewCmdInterface(f))
	cmd.AddCommand(NewCmdLocation(f))
	cmd.AddCommand(NewCmdExec(f))
	cmd.AddCommand(NewCmdUplink(f))
	cmd.AddCommand(NewCmdDatausage(f))
	cmd.AddCommand(NewCmdShadow(f))
	cmd.AddCommand(NewCmdGroup(f))
	cmd.AddCommand(NewCmdConfig(f))
	cmd.AddCommand(NewCmdClient(f))
	cmd.AddCommand(NewCmdAsset(f))

	return cmd
}
