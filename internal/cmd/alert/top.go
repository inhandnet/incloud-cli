package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTop(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top-K alert statistics",
		Long:  "Show top-K devices or alert types ranked by alert count.",
	}

	cmd.AddCommand(NewCmdTopDevices(f))
	cmd.AddCommand(NewCmdTopTypes(f))

	return cmd
}
