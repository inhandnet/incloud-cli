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

type grepDisplayLine struct {
	Line    int
	Text    string
	IsMatch bool
}

type grepDisplayGroup struct {
	DocumentID  string
	SectionID   string
	HeadingPath string
	Lines       []grepDisplayLine
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

			if ctxLines > 0 {
				for i, group := range mergeGrepContextGroups(resp.Matches) {
					if i > 0 {
						fmt.Fprintln(out, c.Gray("--"))
					}
					for _, line := range group.Lines {
						formatted := fmt.Sprintf("%s : %*d : %s", group.HeadingPath, width, line.Line, line.Text)
						if line.IsMatch {
							fmt.Fprintln(out, formatted)
						} else {
							fmt.Fprintln(out, c.Gray(formatted))
						}
					}
				}
			} else {
				for _, m := range resp.Matches {
					fmt.Fprintf(out, "%s : %*d : %s\n", m.HeadingPath, width, m.Line, m.Text)
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

func mergeGrepContextGroups(matches []grepMatch) []grepDisplayGroup {
	groups := make([]grepDisplayGroup, 0, len(matches))
	for _, match := range matches {
		lines := make([]grepDisplayLine, 0, len(match.ContextBefore)+1+len(match.ContextAfter))
		for _, line := range match.ContextBefore {
			lines = append(lines, grepDisplayLine{Line: line.Line, Text: line.Text})
		}
		lines = append(lines, grepDisplayLine{Line: match.Line, Text: match.Text, IsMatch: true})
		for _, line := range match.ContextAfter {
			lines = append(lines, grepDisplayLine{Line: line.Line, Text: line.Text})
		}

		lastIndex := len(groups) - 1
		if lastIndex >= 0 && sameGrepSection(groups[lastIndex], match.DocumentID, match.SectionID) &&
			lines[0].Line <= groups[lastIndex].Lines[len(groups[lastIndex].Lines)-1].Line+1 {
			groups[lastIndex].Lines = mergeGrepDisplayLines(groups[lastIndex].Lines, lines)
			continue
		}
		groups = append(groups, grepDisplayGroup{
			DocumentID:  match.DocumentID,
			SectionID:   match.SectionID,
			HeadingPath: match.HeadingPath,
			Lines:       lines,
		})
	}
	return groups
}

func sameGrepSection(group grepDisplayGroup, documentID, sectionID string) bool {
	return group.DocumentID == documentID && group.SectionID == sectionID
}

func mergeGrepDisplayLines(existing, incoming []grepDisplayLine) []grepDisplayLine {
	for _, line := range incoming {
		insertAt := len(existing)
		for i, current := range existing {
			if current.Line == line.Line {
				if line.IsMatch {
					existing[i] = line
				}
				insertAt = -1
				break
			}
			if current.Line > line.Line {
				insertAt = i
				break
			}
		}
		if insertAt < 0 {
			continue
		}
		existing = append(existing, grepDisplayLine{})
		copy(existing[insertAt+1:], existing[insertAt:])
		existing[insertAt] = line
	}
	return existing
}
