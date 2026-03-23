package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdConfigHistory(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Configuration snapshots",
		Long:  "View, inspect, and restore configuration snapshots for a device.",
	}

	cmd.AddCommand(newCmdConfigHistoryList(f))
	cmd.AddCommand(newCmdConfigHistoryGet(f))
	cmd.AddCommand(newCmdConfigHistoryRestore(f))
	cmd.AddCommand(newCmdConfigHistoryDiff(f))

	return cmd
}
