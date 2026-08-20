package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type grepRequest struct {
	Pattern    string `json:"pattern"`
	DocumentID string `json:"document_id,omitempty"`
	Path       string `json:"path,omitempty"`
	IgnoreCase bool   `json:"ignore_case"`
	Limit      int    `json:"limit,omitempty"`
}

type grepResponse struct {
	Pattern    string      `json:"pattern"`
	MatchCount int         `json:"match_count"`
	Matches    []grepMatch `json:"matches"`
	Truncated  bool        `json:"truncated"`
}

type grepMatch struct {
	DocumentID  string `json:"document_id"`
	SectionID   string `json:"section_id"`
	HeadingPath string `json:"heading_path"`
	Line        int    `json:"line"`
	Text        string `json:"text"`
}

func NewCmdGrep(f *factory.Factory) *cobra.Command {
	var (
		doc        string
		path       string
		ignoreCase bool
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "grep <pattern>",
		Short: "Grep the knowledge base for exact terms",
		Long:  "Locate exact terms (error codes, API names, config keys, versions, protocol names) across the corpus with line numbers. Regex is supported; an invalid regex degrades to a literal substring match. Navigation only — fetch the body with `knowledge read`.",
		Example: `  # Where does an error code appear
  incloud knowledge grep "ERROR 4012"

  # Case-insensitive, inside one document
  incloud knowledge grep IKE --doc 3f2a... -i

  # Regex with a path filter
  incloud knowledge grep "vpn.*(up|down)" --path device_er805`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := grepRequest{
				Pattern:    args[0],
				DocumentID: doc,
				Path:       path,
				IgnoreCase: ignoreCase,
				Limit:      limit,
			}

			body, err := client.Post(agenticBase+"/grep", req)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			var resp grepResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing grep response: %w", err)
			}

			out := f.IO.Out
			errOut := f.IO.ErrOut

			maxLine := 0
			for _, m := range resp.Matches {
				if m.Line > maxLine {
					maxLine = m.Line
				}
			}
			width := len(fmt.Sprintf("%d", maxLine))

			for _, m := range resp.Matches {
				fmt.Fprintf(out, "%s : %*d : %s\n", m.HeadingPath, width, m.Line, m.Text)
			}

			if resp.Truncated {
				fmt.Fprintln(errOut, "Results truncated; raise --limit to see more.")
			}
			if resp.MatchCount == 0 {
				fmt.Fprintln(errOut, "No matches found.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&doc, "doc", "", "Limit search to this document ID")
	cmd.Flags().StringVar(&path, "path", "", "Filter by corpus path prefix (s3_key or filename)")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive match")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max number of matches (1-200)")

	return cmd
}
