package knowledge

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type readRequest struct {
	ChunkID string `json:"chunk_id"`
	Cursor  int    `json:"cursor,omitempty"`
}

type readResponse struct {
	Status     string     `json:"status"`
	Text       string     `json:"text"`
	Source     readSource `json:"source"`
	Truncated  bool       `json:"truncated"`
	NextCursor *int       `json:"next_cursor"`
}

type readSource struct {
	ChunkID     string `json:"chunk_id"`
	Path        string `json:"path"`
	DocTitle    string `json:"doc_title"`
	HeadingPath string `json:"heading_path"`
	URL         string `json:"url"`
}

func NewCmdRead(f *factory.Factory) *cobra.Command {
	var cursor int

	cmd := &cobra.Command{
		Use:   "read <chunk_id>",
		Short: "Read the text of a section",
		Long:  "Fetch the full text of a section by chunk ID (from search or browse). Output is chunked at 12000 characters; continue with --cursor using the returned next_cursor. Chunk IDs change when the documentation is updated; search again if one is not found.",
		Example: `  # Read one section
  incloud knowledge read 1ef0f180...

  # Continue after a 12000-character chunk
  incloud knowledge read 1ef0f180... --cursor 12000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cursor < 0 {
				return fmt.Errorf("--cursor must not be negative")
			}

			body, err := post(f, "/read", readRequest{ChunkID: args[0], Cursor: cursor})
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

			if resp.Status == "empty" {
				fmt.Fprintln(f.IO.ErrOut, "Section not found; chunk IDs change when the documentation is updated, search again.")
				return nil
			}

			out := f.IO.Out
			c := iostreams.NewColorizer(f.IO.TermOutput())

			fmt.Fprint(out, resp.Text)
			if resp.Text != "" && !strings.HasSuffix(resp.Text, "\n") {
				fmt.Fprintln(out)
			}

			meta := fmt.Sprintf("[source: %s > %s]", resp.Source.DocTitle, resp.Source.HeadingPath)
			if resp.Source.URL != "" {
				meta += " " + resp.Source.URL
			}
			fmt.Fprintln(out, c.Gray(meta))

			if resp.Truncated && resp.NextCursor != nil {
				fmt.Fprintf(f.IO.ErrOut, "Output chunked at 12000 characters; continue with --cursor %d.\n", *resp.NextCursor)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&cursor, "cursor", 0, "Continue from a previous response's next_cursor")

	return cmd
}
