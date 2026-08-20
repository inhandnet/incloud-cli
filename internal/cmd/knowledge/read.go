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
// positional argument works for either kind of ID.
type readRequest struct {
	SectionID  string `json:"section_id"`
	DocumentID string `json:"document_id"`
	Offset     *int   `json:"offset,omitempty"`
	Limit      *int   `json:"limit,omitempty"`
}

type readResponse struct {
	Text      string      `json:"text"`
	Source    *readSource `json:"source"`
	Truncated bool        `json:"truncated"`
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
		offset int
		limit  int
	)

	cmd := &cobra.Command{
		Use:   "read <section_id|document_id>",
		Short: "Read the text of a section or a whole document",
		Long:  "Fetch the raw text of a located section, or of a whole document when given a document ID. Output is truncated at 12000 characters; paginate with --offset/--limit.",
		Example: `  # Read one section (section_id from search/grep -o json or browse)
  incloud knowledge read 9c1d...

  # Read a whole document, first 100 lines
  incloud knowledge read 3f2a... --limit 100

  # Continue reading
  incloud knowledge read 3f2a... --offset 100 --limit 100`,
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
			if offset >= 0 {
				req.Offset = &offset
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

			fmt.Fprintln(out, c.Gray(fmt.Sprintf("[source: %s > %s]", resp.Source.Title, resp.Source.HeadingPath)))

			if resp.Truncated {
				fmt.Fprintln(errOut, "Output truncated at 12000 characters; continue with --offset.")
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&offset, "offset", -1, "Start line (>= 0)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max lines (1-2000)")

	return cmd
}
