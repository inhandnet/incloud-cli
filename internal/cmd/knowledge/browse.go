package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type browseRequest struct {
	Product string `json:"product,omitempty"`
	Path    string `json:"path,omitempty"`
	Cursor  int    `json:"cursor,omitempty"`
}

type browseResponse struct {
	Products   []browseProduct  `json:"products"`
	Product    string           `json:"product"`
	Documents  []browseDocument `json:"documents"`
	Path       string           `json:"path"`
	DocTitle   string           `json:"doc_title"`
	Sections   []browseSection  `json:"sections"`
	NextCursor *int             `json:"next_cursor"`
}

type browseProduct struct {
	ProductID   string `json:"product_id"`
	DisplayName string `json:"display_name"`
	Kind        string `json:"kind"`
}

type browseDocument struct {
	Path         string `json:"path"`
	DocTitle     string `json:"doc_title"`
	DocumentType string `json:"document_type"`
}

type browseSection struct {
	ChunkID     string `json:"chunk_id"`
	HeadingPath string `json:"heading_path"`
}

func NewCmdBrowse(f *factory.Factory) *cobra.Command {
	var (
		product string
		cursor  int
	)

	cmd := &cobra.Command{
		Use:   "browse [<document path>]",
		Short: "Browse products and document outlines",
		Long: `Browse the documentation library:

  no arguments      -> all products (product IDs)
  --product <id>    -> product overview: its documents and overview sections
  <document path>   -> section outline of that document, paged with --cursor

Feed chunk IDs into ` + "`knowledge read`" + `.`,
		Example: `  # Which products have documentation
  incloud knowledge browse

  # Documents of one product
  incloud knowledge browse --product DeviceLive

  # Section outline of a document, then the next page
  incloud knowledge browse "docs/zh/ER805/Manuals/用户手册/ER805用户手册_V1.0.md"
  incloud knowledge browse "docs/zh/ER805/Manuals/用户手册/ER805用户手册_V1.0.md" --cursor 10`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := browseRequest{Product: product, Cursor: cursor}
			if len(args) == 1 {
				req.Path = args[0]
			}
			if req.Path != "" && req.Product != "" {
				return fmt.Errorf("give either a document path or --product, not both")
			}
			if req.Path == "" && cmd.Flags().Changed("cursor") {
				return fmt.Errorf("--cursor is only valid with a document path")
			}

			client, err := f.APIClient()
			if err != nil {
				return err
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
			case len(resp.Products) > 0:
				for _, p := range resp.Products {
					fmt.Fprintf(out, "%s %s\n", c.Bold(p.ProductID), c.Gray(fmt.Sprintf("(%s, %s)", p.DisplayName, p.Kind)))
				}
			case resp.Product != "":
				fmt.Fprintln(out, c.Bold(resp.Product))
				for _, d := range resp.Documents {
					fmt.Fprintf(out, "  %s %s\n", d.DocTitle, c.Gray(fmt.Sprintf("[%s] %s", d.DocumentType, d.Path)))
				}
				for _, s := range resp.Sections {
					fmt.Fprintf(out, "  %s %s\n", s.HeadingPath, c.Gray("["+s.ChunkID+"]"))
				}
			case resp.Path != "":
				fmt.Fprintln(out, c.Bold(resp.DocTitle))
				for _, s := range resp.Sections {
					fmt.Fprintf(out, "  %s %s\n", s.HeadingPath, c.Gray("["+s.ChunkID+"]"))
				}
				if resp.NextCursor != nil {
					fmt.Fprintf(errOut, "More sections; continue with --cursor %d.\n", *resp.NextCursor)
				}
			default:
				fmt.Fprintln(errOut, "Nothing found.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&product, "product", "", "Show the overview of a product (e.g. DeviceLive, ER805)")
	cmd.Flags().IntVar(&cursor, "cursor", 0, "Continue a document outline from a previous response's next_cursor")

	return cmd
}
