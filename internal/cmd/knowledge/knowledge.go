package knowledge

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// agenticBase is the copilot agentic-search REST prefix backing this command group.
const agenticBase = "/api/v1/knowledge/agentic"

func NewCmdKnowledge(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "knowledge",
		Short: "Search the knowledge base",
		Long:  "Search and read device documentation: search for candidates, browse document outlines, grep for exact terms, read section text.",
	}

	cmd.AddCommand(NewCmdSearch(f))
	cmd.AddCommand(NewCmdBrowse(f))
	cmd.AddCommand(NewCmdGrep(f))
	cmd.AddCommand(NewCmdRead(f))

	return cmd
}
