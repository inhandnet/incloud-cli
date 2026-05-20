package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdApp(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage device applications",
		Long:  "List, start, stop, and restart applications running on edge devices.",
	}

	cmd.AddCommand(newCmdAppList(f))
	cmd.AddCommand(newCmdAppStart(f))
	cmd.AddCommand(newCmdAppStop(f))
	cmd.AddCommand(newCmdAppRestart(f))

	return cmd
}
