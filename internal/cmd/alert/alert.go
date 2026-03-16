package alert

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdAlert(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage alerts",
		Long:  "List, view, acknowledge alerts and check alert statistics on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdAck(f))
	cmd.AddCommand(NewCmdExport(f))

	return cmd
}
