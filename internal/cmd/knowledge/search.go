package knowledge

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type searchRequest struct {
	Query   string `json:"query"`
	Model   string `json:"model,omitempty"`
	Rewrite bool   `json:"rewrite,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

type searchResponse struct {
	Query            string         `json:"query"`
	RewrittenQueries []string       `json:"rewritten_queries"`
	Results          []searchResult `json:"results"`
}

type searchResult struct {
	Content      string `json:"content"`
	Source       string `json:"source"`
	Heading      string `json:"heading"`
	DocumentType string `json:"document_type"`
	Model        string `json:"model"`
}

var collapseWS = regexp.MustCompile(`\s+`)

func NewCmdSearch(f *factory.Factory) *cobra.Command {
	var (
		model   string
		rewrite bool
		limit   int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the knowledge base",
		Long:  "Search device documentation and return matching results.",
		Example: `  # Search for configuration guides
  incloud knowledge search "how to configure VPN"

  # Filter by device model
  incloud knowledge search "factory reset" --model IR915L

  # Enable query rewriting for better results
  incloud knowledge search "VPN setup" --rewrite

  # Limit results and output as JSON
  incloud knowledge search "firewall rules" --limit 3 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := searchRequest{
				Query:   args[0],
				Model:   model,
				Rewrite: rewrite,
				Limit:   limit,
			}

			body, err := client.Post("/api/v1/knowledge/search", req)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			var resp searchResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing search response: %w", err)
			}

			out := f.IO.Out
			c := iostreams.NewColorizer(f.IO.TermOutput())

			for i, r := range resp.Results {
				if i > 0 {
					fmt.Fprintln(out)
				}
				meta := r.Source
				if r.Model != "" && r.Model != "default" {
					meta = fmt.Sprintf("[%s] %s", strings.ToUpper(r.Model), meta)
				}
				fmt.Fprintln(out, c.Bold(r.Heading))
				fmt.Fprintln(out, c.Gray(meta))
				snippet := collapseWS.ReplaceAllString(strings.TrimSpace(r.Content), " ")
				fmt.Fprintln(out, snippet)
			}

			if len(resp.Results) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "No results found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Filter by device model (e.g. IR915L)")
	cmd.Flags().BoolVar(&rewrite, "rewrite", false, "Enable LLM query rewriting")
	cmd.Flags().IntVar(&limit, "limit", 6, "Max number of results (1-20)")

	return cmd
}
