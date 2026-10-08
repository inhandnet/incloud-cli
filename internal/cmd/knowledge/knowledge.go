package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// agenticBase is the copilot knowledge REST prefix backing this command group.
const agenticBase = "/api/v1/knowledge/agentic"

func NewCmdKnowledge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Search the knowledge base",
		Long:  "Search and read InHand product documentation: search for candidate sections, browse products and document outlines, read section text.",
	}

	cmd.AddCommand(NewCmdSearch(f))
	cmd.AddCommand(NewCmdBrowse(f))
	cmd.AddCommand(NewCmdRead(f))

	return cmd
}
