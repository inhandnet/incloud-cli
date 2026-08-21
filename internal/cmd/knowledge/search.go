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
	Query string `json:"query"`
	Model string `json:"model,omitempty"`
	Path  string `json:"path,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type searchResponse struct {
	Status  string         `json:"status"`
	Results []searchResult `json:"results"`
}

type searchResult struct {
	Source       string  `json:"source"`
	DocumentID   string  `json:"document_id"`
	SectionID    string  `json:"section_id"`
	HeadingPath  string  `json:"heading_path"`
	DocumentType string  `json:"document_type"`
	Model        string  `json:"model"`
	FromFallback bool    `json:"from_fallback"`
	Score        float64 `json:"score"`
	Snippet      string  `json:"snippet"`
}

var collapseWS = regexp.MustCompile(`\s+`)

func NewCmdSearch(f *factory.Factory) *cobra.Command {
	var (
		model string
		path  string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the knowledge base",
		Long:  "Search device documentation and return addressable section candidates. Snippets are for picking a section only; fetch the body with `knowledge read`.",
		Example: `  # Search for configuration guides
  incloud knowledge search "how to configure VPN"

  # Filter by device model
  incloud knowledge search "factory reset" --model IR915L

  # Filter by corpus path prefix (s3_key or filename)
  incloud knowledge search "firewall rules" --path device_

  # Limit results and output as JSON (full fields incl. ids)
  incloud knowledge search "firewall rules" --limit 3 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := searchRequest{
				Query: args[0],
				Model: model,
				Path:  path,
				Limit: limit,
			}

			body, err := client.Post(agenticBase+"/search", req)
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
			errOut := f.IO.ErrOut
			c := iostreams.NewColorizer(f.IO.TermOutput())

			switch resp.Status {
			case "kb_not_ready":
				fmt.Fprintln(errOut, "Knowledge base is not ready. Try again later.")
				return nil
			case "failed":
				fmt.Fprintln(errOut, "Search failed on the server. Try again later.")
				return nil
			case "model_not_found":
				fmt.Fprintln(errOut, "No dedicated docs for this model; showing fallback results from the whole corpus.")
			}

			for i := range resp.Results {
				if i > 0 {
					fmt.Fprintln(out)
				}
				r := &resp.Results[i]
				meta := r.Source
				if r.Model != "" && r.Model != "default" {
					meta = fmt.Sprintf("[%s] %s", strings.ToUpper(r.Model), meta)
				}
				if r.FromFallback {
					meta += " (fallback)"
				}
				fmt.Fprintln(out, c.Bold(r.HeadingPath))
				fmt.Fprintln(out, c.Gray(meta))
				fmt.Fprintln(out, collapseWS.ReplaceAllString(strings.TrimSpace(r.Snippet), " "))
			}

			if len(resp.Results) == 0 {
				fmt.Fprintln(errOut, "No results found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Filter by device model (e.g. IR915L)")
	cmd.Flags().StringVar(&path, "path", "", "Filter by corpus path prefix (s3_key or filename)")
	cmd.Flags().IntVar(&limit, "limit", 10, "Max number of results (1-50)")

	return cmd
}
