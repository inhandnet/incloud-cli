package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdDevice(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "device",
		Aliases: []string{"dev"},
		Short:   "Manage devices in connector networks",
	}

	cmd.AddCommand(newCmdDeviceList(f))
	cmd.AddCommand(newCmdDeviceListAll(f))
	cmd.AddCommand(newCmdDeviceAdd(f))
	cmd.AddCommand(newCmdDeviceUpdate(f))
	cmd.AddCommand(newCmdDeviceDelete(f))
	cmd.AddCommand(newCmdDeviceCandidates(f))

	return cmd
}
