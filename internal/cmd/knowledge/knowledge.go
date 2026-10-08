package knowledge

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

// knowledgeBase is the copilot knowledge REST prefix backing this command group.
const knowledgeBase = "/api/v1/knowledge/v2"

// post sends a knowledge request; a 404 on the endpoint itself means the
// server predates the v2 API, which deserves a clearer message than the raw body.
func post(f *factory.Factory, endpoint string, body any) ([]byte, error) {
	client, err := f.APIClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Post(knowledgeBase+endpoint, body)
	var httpErr *api.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("HTTP 404: the server does not provide %s%s (knowledge v2 API); the copilot backend is older than this CLI", knowledgeBase, endpoint)
	}
	return resp, err
}

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
