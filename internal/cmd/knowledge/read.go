package knowledge

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// readRequest carries the same ID as both fields: the server resolves
// section IDs first and falls back to whole-document reads, so a single
// positional argument works for either kind of ID. Line parameters are
// 1-based, relative to the section body (line 1 is the section heading).
type readRequest struct {
	SectionID   string  `json:"section_id"`
	DocumentID  string  `json:"document_id"`
	Mode        *string `json:"mode,omitempty"`
	LineStart   *int    `json:"line_start,omitempty"`
	LineEnd     *int    `json:"line_end,omitempty"`
	AroundLine  *int    `json:"around_line,omitempty"`
	Before      *int    `json:"before,omitempty"`
	After       *int    `json:"after,omitempty"`
	Limit       *int    `json:"limit,omitempty"`
}

type readResponse struct {
	Text       string      `json:"text"`
	Source     *readSource `json:"source"`
	Truncated  bool        `json:"truncated"`
	TotalLines int         `json:"total_lines"`
}

type readSource struct {
	DocumentID  string `json:"document_id"`
	SectionID   string `json:"section_id"`
	S3Key       string `json:"s3_key"`
	Title       string `json:"title"`
	HeadingPath string `json:"heading_path"`
}

func NewCmdRead(f *factory.Factory) *cobra.Command {
	var (
		mode      string
		lineStart int
		lineEnd   int
		around    int
		before    int
		after     int
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "read <section_id|document_id>",
		Short: "Read the text of a section or a whole document",
		Long:  "Fetch the raw text of a located section, or of a whole document when given a document ID. Line numbers are 1-based, relative to the section body (line 1 is the section heading itself); a grep hit's (section_id, line) feeds directly into --mode around. Output is truncated at 12000 characters in every mode. Modes: full (default), range (--line-start/--line-end, inclusive), around (--around ± --before/--after, defaults 20/20), head/tail (--limit lines).",
		Example: `  # Read one section (section_id from search/grep -o json or browse)
  incloud knowledge read 9c1d...

  # Lines 10-50 (1-based, inclusive)
  incloud knowledge read 9c1d... --mode range --line-start 10 --line-end 50

  # Around a grep hit at line 186
  incloud knowledge read 9c1d... --mode around --around 186 --before 30 --after 50

  # First / last N lines
  incloud knowledge read 3f2a... --mode head --limit 100
  incloud knowledge read 3f2a... --mode tail --limit 20`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := readRequest{
				SectionID:  args[0],
				DocumentID: args[0],
			}
			if mode != "" {
				req.Mode = &mode
			}
			if lineStart >= 0 {
				req.LineStart = &lineStart
			}
			if lineEnd >= 0 {
				req.LineEnd = &lineEnd
			}
			if around >= 0 {
				req.AroundLine = &around
			}
			if before >= 0 {
				req.Before = &before
			}
			if after >= 0 {
				req.After = &after
			}
			if limit > 0 {
				req.Limit = &limit
			}

			body, err := client.Post(agenticBase+"/read", req)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			var resp readResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing read response: %w", err)
			}
			if resp.Source == nil {
				fmt.Fprintln(f.IO.ErrOut, "Section/document not found or knowledge base not ready.")
				return nil
			}

			out := f.IO.Out
			errOut := f.IO.ErrOut
			c := iostreams.NewColorizer(f.IO.TermOutput())

			fmt.Fprint(out, resp.Text)
			if resp.Text != "" && !strings.HasSuffix(resp.Text, "\n") {
				fmt.Fprintln(out)
			}

			meta := fmt.Sprintf("[source: %s > %s]", resp.Source.Title, resp.Source.HeadingPath)
			if resp.TotalLines > 0 {
				meta += fmt.Sprintf(" · %d lines total", resp.TotalLines)
			}
			fmt.Fprintln(out, c.Gray(meta))

			if resp.Truncated {
				fmt.Fprintln(errOut, "Output truncated at 12000 characters; continue reading with --mode range.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "", "Read mode: full (default), range, around, head, tail")
	cmd.Flags().IntVar(&lineStart, "line-start", -1, "First line, 1-based inclusive (mode=range; alone = to the end)")
	cmd.Flags().IntVar(&lineEnd, "line-end", -1, "Last line, 1-based inclusive (mode=range; alone = from line 1)")
	cmd.Flags().IntVar(&around, "around", -1, "Center line for mode=around (1-based)")
	cmd.Flags().IntVar(&before, "before", -1, "Lines before --around (mode=around, default 20)")
	cmd.Flags().IntVar(&after, "after", -1, "Lines after --around (mode=around, default 20)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of lines (mode=head/tail, 1-2000)")

	return cmd
}
