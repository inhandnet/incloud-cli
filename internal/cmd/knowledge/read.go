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
	SectionID  string  `json:"section_id"`
	DocumentID string  `json:"document_id"`
	Mode       *string `json:"mode,omitempty"`
	LineStart  *int    `json:"line_start,omitempty"`
	LineEnd    *int    `json:"line_end,omitempty"`
	AroundLine *int    `json:"around_line,omitempty"`
	Before     *int    `json:"before,omitempty"`
	After      *int    `json:"after,omitempty"`
	Limit      *int    `json:"limit,omitempty"`
	Cursor     *string `json:"cursor,omitempty"`
}

type readResponse struct {
	Text       string      `json:"text"`
	Source     *readSource `json:"source"`
	Truncated  bool        `json:"truncated"`
	TotalLines int         `json:"total_lines"`
	NextCursor string      `json:"next_cursor"`
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
		cursor    string
	)

	cmd := &cobra.Command{
		Use:   "read <section_id|document_id>",
		Short: "Read the text of a section or a whole document",
		Long:  "Fetch the raw text of a located section, or of a whole document when given a document ID. Line numbers are 1-based, relative to the section body (line 1 is the section heading itself); a grep hit's (section_id, line) feeds directly into --mode around. Output is chunked at 12000 characters in every mode; continue from the exact character with --cursor using the returned next_cursor. Modes: full (default), range (--line-start/--line-end, inclusive), around (--around ± --before/--after, defaults 20/20), head/tail (--limit lines).",
		Example: `  # Read one section (section_id from search/grep -o json or browse)
  incloud knowledge read 9c1d...

  # Lines 10-50 (1-based, inclusive)
  incloud knowledge read 9c1d... --mode range --line-start 10 --line-end 50

  # Around a grep hit at line 186
  incloud knowledge read 9c1d... --mode around --around 186 --before 30 --after 50

  # First / last N lines
  incloud knowledge read 3f2a... --mode head --limit 100
  incloud knowledge read 3f2a... --mode tail --limit 20

  # Continue after a 12000-character chunk (works inside a single long line)
  incloud knowledge read 9c1d... --cursor '<next_cursor>'`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validateReadFlags(cmd, mode, lineStart, lineEnd, around, before, after, limit, cursor)
		},
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
			if cmd.Flags().Changed("line-start") {
				req.LineStart = &lineStart
			}
			if cmd.Flags().Changed("line-end") {
				req.LineEnd = &lineEnd
			}
			if cmd.Flags().Changed("around") {
				req.AroundLine = &around
			}
			if cmd.Flags().Changed("before") {
				req.Before = &before
			}
			if cmd.Flags().Changed("after") {
				req.After = &after
			}
			if cmd.Flags().Changed("limit") {
				req.Limit = &limit
			}
			if cmd.Flags().Changed("cursor") {
				req.Cursor = &cursor
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
				if resp.NextCursor != "" {
					fmt.Fprintf(errOut, "Output chunked at 12000 characters; continue with --cursor '%s'.\n", resp.NextCursor)
				} else {
					fmt.Fprintln(errOut, "Output truncated at 12000 characters; continue reading with --mode range.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&mode, "mode", "", "Read mode: full (default), range, around, head, tail")
	cmd.Flags().IntVar(&lineStart, "line-start", 0, "First line, 1-based inclusive (mode=range; alone = to the end)")
	cmd.Flags().IntVar(&lineEnd, "line-end", 0, "Last line, 1-based inclusive (mode=range; alone = from line 1)")
	cmd.Flags().IntVar(&around, "around", 0, "Center line for mode=around (1-based)")
	cmd.Flags().IntVar(&before, "before", 20, "Lines before --around (mode=around)")
	cmd.Flags().IntVar(&after, "after", 20, "Lines after --around (mode=around)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Number of lines (mode=head/tail, 1-2000)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Continue from a previous response's next_cursor")

	return cmd
}

func validateReadFlags(
	cmd *cobra.Command,
	mode string,
	lineStart, lineEnd, around, before, after, limit int,
	cursor string,
) error {
	if mode == "" {
		mode = "full"
	}
	validModes := map[string]bool{
		"full": true, "range": true, "around": true, "head": true, "tail": true,
	}
	if !validModes[mode] {
		return fmt.Errorf("invalid --mode %q: must be full, range, around, head, or tail", mode)
	}

	changed := func(name string) bool { return cmd.Flags().Changed(name) }
	lineFlags := []string{"line-start", "line-end", "around", "before", "after", "limit"}
	if changed("cursor") {
		if cursor == "" {
			return fmt.Errorf("--cursor cannot be empty")
		}
		if changed("mode") {
			return fmt.Errorf("--mode is not valid with --cursor")
		}
		for _, name := range lineFlags {
			if changed(name) {
				return fmt.Errorf("--%s is not valid with --cursor", name)
			}
		}
		return nil
	}
	if changed("line-start") && lineStart < 1 {
		return fmt.Errorf("--line-start must be at least 1")
	}
	if changed("line-end") && lineEnd < 1 {
		return fmt.Errorf("--line-end must be at least 1")
	}
	if changed("around") && around < 1 {
		return fmt.Errorf("--around must be at least 1")
	}
	if changed("before") && (before < 0 || before > 2000) {
		return fmt.Errorf("--before must be between 0 and 2000")
	}
	if changed("after") && (after < 0 || after > 2000) {
		return fmt.Errorf("--after must be between 0 and 2000")
	}
	if changed("limit") && (limit < 1 || limit > 2000) {
		return fmt.Errorf("--limit must be between 1 and 2000")
	}

	rejectExcept := func(allowed map[string]bool) error {
		for _, name := range lineFlags {
			if changed(name) && !allowed[name] {
				return fmt.Errorf("--%s is not valid with --mode %s", name, mode)
			}
		}
		return nil
	}

	switch mode {
	case "full":
		return rejectExcept(nil)
	case "range":
		if !changed("line-start") && !changed("line-end") {
			return fmt.Errorf("--mode range requires --line-start or --line-end")
		}
		if err := rejectExcept(map[string]bool{"line-start": true, "line-end": true}); err != nil {
			return err
		}
		if changed("line-start") && changed("line-end") && lineStart > lineEnd {
			return fmt.Errorf("--line-start cannot be greater than --line-end")
		}
	case "around":
		if !changed("around") {
			return fmt.Errorf("--mode around requires --around")
		}
		return rejectExcept(map[string]bool{"around": true, "before": true, "after": true})
	case "head", "tail":
		if !changed("limit") {
			return fmt.Errorf("--mode %s requires --limit", mode)
		}
		return rejectExcept(map[string]bool{"limit": true})
	}

	return nil
}
