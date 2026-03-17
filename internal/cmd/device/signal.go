package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdSignal(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signal",
		Short: "Device signal quality",
		Long:  "View, export, and summarize device signal quality metrics (RSRP, RSRQ, SINR, etc.).",
	}

	cmd.AddCommand(newCmdSignalList(f))
	cmd.AddCommand(newCmdSignalExport(f))

	return cmd
}
