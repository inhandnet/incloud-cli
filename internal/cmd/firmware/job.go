package firmware

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdJob(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage OTA firmware upgrade jobs",
		Long:  "Create and inspect OTA firmware upgrade jobs.",
	}

	cmd.AddCommand(NewCmdJobList(f))
	cmd.AddCommand(NewCmdJobCreate(f))
	cmd.AddCommand(NewCmdJobCancel(f))
	cmd.AddCommand(NewCmdJobExecutions(f))

	return cmd
}
