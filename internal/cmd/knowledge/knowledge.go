package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdKnowledge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Search and query the knowledge base",
		Long:  "Search device documentation and ask questions powered by AI.",
	}

	cmd.AddCommand(NewCmdSearch(f))
	cmd.AddCommand(NewCmdAsk(f))

	return cmd
}
