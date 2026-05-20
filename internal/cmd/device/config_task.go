package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdConfigTask(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage CLI configuration tasks",
		Long:  "Create and view CLI configuration tasks that push configuration commands to edge devices.",
	}

	cmd.AddCommand(newCmdConfigTaskCreate(f))
	cmd.AddCommand(newCmdConfigTaskGet(f))

	return cmd
}
