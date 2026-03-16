package activity

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdActivity(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity",
		Short: "View activity logs",
		Long:  "Query audit activity logs on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))

	return cmd
}
