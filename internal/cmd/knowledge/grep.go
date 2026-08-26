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
	Fixed      bool   `json:"fixed,omitempty"`
	Context    int    `json:"context,omitempty"`
}

type grepResponse struct {
	Pattern    string      `json:"pattern"`
	MatchCount int         `json:"match_count"`
	Matches    []grepMatch `json:"matches"`
	Truncated  bool        `json:"truncated"`
}

type grepMatch struct {
	DocumentID    string            `json:"document_id"`
	SectionID     string            `json:"section_id"`
	HeadingPath   string            `json:"heading_path"`
	Line          int               `json:"line"`
	Text          string            `json:"text"`
	ContextBefore []grepContextLine `json:"context_before"`
	ContextAfter  []grepContextLine `json:"context_after"`
}

type grepContextLine struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

func NewCmdGrep(f *factory.Factory) *cobra.Command {
	var (
		doc        string
		path       string
		ignoreCase bool
		limit      int
		fixed      bool
		ctxLines   int
	)

	cmd := &cobra.Command{
		Use:   "grep <pattern>",
		Short: "Grep the knowledge base for exact terms",
		Long:  "Locate exact terms (error codes, API names, config keys, versions, protocol names) across the corpus with line numbers. Regex is supported; an invalid regex degrades to a literal substring match. Use -F to treat the pattern as a fixed string (recommended for versions like 5.1, where the regex dot would also match 571). Use -C to show surrounding context lines. Lines are 1-based, relative to the section body (line 1 is the section heading itself), so a hit's (section_id, line) feeds directly into `knowledge read --mode around`. Navigation only — fetch the body with `knowledge read`.",
		Example: `  # Where does an error code appear
  incloud knowledge grep "ERROR 4012"

  # Exact version string, not regex (5.1 would otherwise also match 571)
  incloud knowledge grep "5.1" -F

  # Case-insensitive with context, inside one document
  incloud knowledge grep IKE --doc 3f2a... -i -C 3

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
				Fixed:      fixed,
				Context:    ctxLines,
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
			c := iostreams.NewColorizer(f.IO.TermOutput())

			maxLine := 0
			for _, m := range resp.Matches {
				if m.Line > maxLine {
					maxLine = m.Line
				}
				for _, cl := range append(m.ContextBefore, m.ContextAfter...) {
					if cl.Line > maxLine {
						maxLine = cl.Line
					}
				}
			}
			width := len(fmt.Sprintf("%d", maxLine))

			for i, m := range resp.Matches {
				if i > 0 && ctxLines > 0 {
					fmt.Fprintln(out, c.Gray("--"))
				}
				for _, cl := range m.ContextBefore {
					fmt.Fprintln(out, c.Gray(fmt.Sprintf("%s : %*d : %s", m.HeadingPath, width, cl.Line, cl.Text)))
				}
				fmt.Fprintf(out, "%s : %*d : %s\n", m.HeadingPath, width, m.Line, m.Text)
				for _, cl := range m.ContextAfter {
					fmt.Fprintln(out, c.Gray(fmt.Sprintf("%s : %*d : %s", m.HeadingPath, width, cl.Line, cl.Text)))
				}
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
	cmd.Flags().BoolVarP(&fixed, "fixed", "F", false, "Treat pattern as a fixed string, not regex (grep -F)")
	cmd.Flags().IntVarP(&ctxLines, "context", "C", 0, "Show N context lines around each match (grep -C, 0 = off)")

	return cmd
}
