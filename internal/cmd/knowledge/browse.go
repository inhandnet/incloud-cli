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
	Product string `json:"product,omitempty"`
	Path    string `json:"path,omitempty"`
}

type browseResponse struct {
	Status   string          `json:"status"`
	Message  string          `json:"message"`
	Products []browseProduct `json:"products"`
	Product  string          `json:"product"`
	Sections []browseSection `json:"sections"`
}

type browseProduct struct {
	ProductID   string `json:"product_id"`
	DisplayName string `json:"display_name"`
}

// browseSection covers both shapes: product overview sections carry path and
// heading_path, document outline sections carry title and level.
type browseSection struct {
	ChunkID     string `json:"chunk_id"`
	Path        string `json:"path"`
	HeadingPath string `json:"heading_path"`
	Title       string `json:"title"`
	Level       int    `json:"level"`
}

func NewCmdBrowse(f *factory.Factory) *cobra.Command {
	var product string

	cmd := &cobra.Command{
		Use:   "browse [<document path>]",
		Short: "Browse products and document outlines",
		Long: `Browse the documentation library:

  no arguments      -> all products (product IDs)
  --product <id>    -> product overview sections
  <document path>   -> full section outline of that document

Feed chunk IDs into ` + "`knowledge read`" + `.`,
		Example: `  # Which products have documentation
  incloud knowledge browse

  # Overview of one product
  incloud knowledge browse --product DeviceLive

  # Full section outline of a document
  incloud knowledge browse "docs/zh/ER805/Manuals/用户手册/ER805用户手册_V1.0.md"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req := browseRequest{Product: product}
			if len(args) == 1 {
				req.Path = args[0]
			}
			if req.Path != "" && req.Product != "" {
				return fmt.Errorf("give either a document path or --product, not both")
			}

			body, err := post(f, "/browse", req)
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
			case resp.Status == "failed":
				fmt.Fprintf(errOut, "Browse failed: %s\n", resp.Message)
			case resp.Status == "empty":
				fmt.Fprintln(errOut, "Nothing found.")
			case req.Path != "":
				for _, s := range resp.Sections {
					indent := strings.Repeat("  ", max(s.Level-1, 0))
					fmt.Fprintf(out, "%s%s %s\n", indent, s.Title, c.Gray("["+s.ChunkID+"]"))
				}
			case req.Product != "":
				fmt.Fprintln(out, c.Bold(resp.Product))
				for _, s := range resp.Sections {
					fmt.Fprintf(out, "  %s %s\n", s.HeadingPath, c.Gray(fmt.Sprintf("%s [%s]", s.Path, s.ChunkID)))
				}
			default:
				for _, p := range resp.Products {
					line := c.Bold(p.ProductID)
					if p.DisplayName != "" && p.DisplayName != p.ProductID {
						line += " " + c.Gray("("+p.DisplayName+")")
					}
					fmt.Fprintln(out, line)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&product, "product", "", "Show the overview of a product (e.g. DeviceLive, ER805)")

	return cmd
}
