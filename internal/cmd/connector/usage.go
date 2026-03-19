package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdUsage(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Connector traffic usage statistics",
	}

	cmd.AddCommand(newCmdUsageStats(f))
	cmd.AddCommand(newCmdUsageTrend(f))
	cmd.AddCommand(newCmdUsageTopK(f))

	return cmd
}
