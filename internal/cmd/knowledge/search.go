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
	Limit int    `json:"limit,omitempty"`
}

type searchResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Results []searchResult `json:"results"`
}

type searchResult struct {
	ChunkID     string   `json:"chunk_id"`
	Path        string   `json:"path"`
	DocTitle    string   `json:"doc_title"`
	HeadingPath string   `json:"heading_path"`
	ProductIDs  []string `json:"product_ids"`
	Score       float64  `json:"score"`
	Snippet     string   `json:"snippet"`
}

var collapseWS = regexp.MustCompile(`\s+`)

func NewCmdSearch(f *factory.Factory) *cobra.Command {
	var (
		model string
		limit int
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the knowledge base",
		Long:  "Keyword search over product documentation; returns candidate sections with chunk IDs. Matching is lexical: include the product or model name and keep queries to a few keywords. Snippets are for picking a section only; fetch the body with `knowledge read`.",
		Example: `  # Search with product name in the query
  incloud knowledge search "DeviceLive 添加设备"

  # Prefix the query with a model
  incloud knowledge search "恢复出厂设置" --model ER805

  # Limit results and output as JSON (full fields incl. chunk IDs)
  incloud knowledge search "IPSec VPN" --limit 3 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := post(f, "/search", searchRequest{
				Query: args[0],
				Model: model,
				Limit: limit,
			})
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

			if resp.Status == "failed" {
				fmt.Fprintf(errOut, "Search failed: %s\n", resp.Message)
				return nil
			}

			for i := range resp.Results {
				if i > 0 {
					fmt.Fprintln(out)
				}
				r := &resp.Results[i]
				meta := r.Path
				if len(r.ProductIDs) > 0 {
					meta = fmt.Sprintf("[%s] %s", strings.Join(r.ProductIDs, ","), meta)
				}
				fmt.Fprintln(out, c.Bold(r.HeadingPath))
				fmt.Fprintln(out, c.Gray(fmt.Sprintf("%s [%s]", meta, r.ChunkID)))
				fmt.Fprintln(out, collapseWS.ReplaceAllString(strings.TrimSpace(r.Snippet), " "))
			}

			if len(resp.Results) == 0 {
				fmt.Fprintln(errOut, "No results found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&model, "model", "", "Product ID or model to add to the query (e.g. ER805, DeviceLive)")
	cmd.Flags().IntVar(&limit, "limit", 5, "Max number of results (1-10)")

	return cmd
}
