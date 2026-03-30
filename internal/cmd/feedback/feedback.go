package feedback

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdFeedback(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Submit feedback about the InCloud platform",
	}

	cmd.AddCommand(NewCmdFeedbackCreate(f))
	cmd.AddCommand(NewCmdFeedbackDownload(f))
	cmd.AddCommand(NewCmdFeedbackList(f))
	cmd.AddCommand(NewCmdFeedbackResolve(f))

	return cmd
}
