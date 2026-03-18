package firmware

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page    int
	Limit   int
	Sort    string
	Product string
	Module  string
	Status  string
	Fields  []string
}

var defaultListFields = []string{"_id", "product", "version", "status", "latest", "order", "publishedAt"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List firmwares",
		Long:    "List firmwares on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List firmwares with default pagination
  incloud firmware list

  # Filter by product
  incloud firmware list --product IR915L

  # Filter by status
  incloud firmware list --status published

  # Paginate
  incloud firmware list --page 2 --limit 50

  # Select fields
  incloud firmware list -f product -f version -f latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Product != "" {
				q.Set("product", opts.Product)
			}
			if opts.Module != "" {
				q.Set("module", opts.Module)
			}
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/firmwares", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Product, "product", "", "Filter by product name")
	cmd.Flags().StringVar(&opts.Module, "module", "", "Filter by module name")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (published|unpublished|deprecated)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
