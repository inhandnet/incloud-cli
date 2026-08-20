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
	DocumentID string  `json:"document_id"`
	SectionID  *string `json:"section_id,omitempty"`
}

type browseResponse struct {
	DocumentID string       `json:"document_id"`
	Title      string       `json:"title"`
	Nodes      []browseNode `json:"nodes"`
}

type browseNode struct {
	SectionID  string `json:"section_id"`
	Title      string `json:"title"`
	Level      int    `json:"level"`
	CharCount  int    `json:"char_count"`
	ChildCount int    `json:"child_count"`
}

func NewCmdBrowse(f *factory.Factory) *cobra.Command {
	var section string

	cmd := &cobra.Command{
		Use:   "browse <document_id>",
		Short: "Browse the section outline of a document",
		Long:  "List the section outline (table of contents) of a knowledge document, or the subtree rooted at a section. Feed section IDs into `knowledge read`.",
		Example: `  # Outline of a document (document_id comes from knowledge search -o json)
  incloud knowledge browse 3f2a...

  # Only the subtree under one section
  incloud knowledge browse 3f2a... --section 9c1d...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			req := browseRequest{DocumentID: args[0]}
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
			if resp.Title == "" {
				fmt.Fprintln(f.IO.ErrOut, "Document not found or knowledge base not ready.")
				return nil
			}

			out := f.IO.Out
			c := iostreams.NewColorizer(f.IO.TermOutput())

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
			return nil
		},
	}

	cmd.Flags().StringVar(&section, "section", "", "Limit output to the subtree rooted at this section ID")

	return cmd
}
