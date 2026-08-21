package knowledge

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type browseRequest struct {
	Path      string  `json:"path,omitempty"`
	SectionID *string `json:"section_id,omitempty"`
}

type browseResponse struct {
	DocumentID string           `json:"document_id"`
	Title      string           `json:"title"`
	Nodes      []browseNode     `json:"nodes"`
	Documents  []browseDocument `json:"documents"`
}

type browseNode struct {
	SectionID  string `json:"section_id"`
	Title      string `json:"title"`
	Level      int    `json:"level"`
	CharCount  int    `json:"char_count"`
	ChildCount int    `json:"child_count"`
}

type browseDocument struct {
	DocumentID   string `json:"document_id"`
	Title        string `json:"title"`
	Source       string `json:"source"`
	S3Key        string `json:"s3_key"`
	DocumentType string `json:"document_type"`
	Model        string `json:"model"`
	Region       string `json:"region"`
	SectionCount int    `json:"section_count"`
	CharCount    int    `json:"char_count"`
}

func NewCmdBrowse(f *factory.Factory) *cobra.Command {
	var section string

	cmd := &cobra.Command{
		Use:   "browse [<path>]",
		Short: "Browse the knowledge base like a filesystem",
		Long: `Browse the knowledge base by corpus path, like ls on a virtual filesystem:

  no path        -> catalog of all documents in the corpus
  path prefix    -> catalog filtered to matching documents (s3_key or filename prefix)
  unique match   -> section outline of that document (--section to open a subtree)

Feed section IDs into ` + "`knowledge read`" + `.`,
		Example: `  # What documents are in the corpus
  incloud knowledge browse

  # Narrow by prefix (model filters work via filename, e.g. device_er805*)
  incloud knowledge browse device_

  # Open one document's outline, then a subtree
  incloud knowledge browse device_er805_用户手册
  incloud knowledge browse device_er805_用户手册 --section 9c1d...`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := browseRequest{}
			if len(args) == 1 {
				req.Path = args[0]
			}
			if section != "" {
				req.SectionID = &section
			}

			body, err := client.Post(agenticBase+"/browse", req)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output != "table" {
				return iostreams.FormatOutput(body, f.IO, output)
			}

			var resp browseResponse
			if err := json.Unmarshal(body, &resp); err != nil {
				return fmt.Errorf("parsing browse response: %w", err)
			}

			out := f.IO.Out
			errOut := f.IO.ErrOut
			c := iostreams.NewColorizer(f.IO.TermOutput())

			switch {
			case len(resp.Documents) > 0:
				for i := range resp.Documents {
					if i > 0 {
						fmt.Fprintln(out)
					}
					d := &resp.Documents[i]
					meta := fmt.Sprintf("%s · %d sections · %d chars",
						d.Source, d.SectionCount, d.CharCount)
					if d.Model != "" && d.Model != "default" {
						meta = fmt.Sprintf("[%s] %s", strings.ToUpper(d.Model), meta)
					}
					fmt.Fprintln(out, c.Bold(d.Title))
					fmt.Fprintln(out, c.Gray(fmt.Sprintf("%s [%s]", meta, d.DocumentID)))
				}
			case resp.Title != "":
				fmt.Fprintln(out, c.Bold(resp.Title))
				for _, n := range resp.Nodes {
					indent := strings.Repeat("  ", max(n.Level-1, 0))
					fmt.Fprintf(out, "%s%s %s\n",
						indent,
						n.Title,
						c.Gray(fmt.Sprintf("(%d chars, %d children) [%s]",
							n.CharCount, n.ChildCount, n.SectionID)),
					)
				}
			default:
				fmt.Fprintln(errOut, "Nothing found at this path (or knowledge base not ready).")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&section, "section", "", "Open the subtree rooted at this section ID (unique path match only)")

	return cmd
}
