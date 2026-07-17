package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdKnowledge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Search the knowledge base",
		Long:  "Search device documentation.",
	}

	cmd.AddCommand(NewCmdSearch(f))

	return cmd
}
