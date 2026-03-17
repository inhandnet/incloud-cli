package product

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page   int
	Limit  int
	Sort   string
	Name   string
	Type   string
	Status string
	Fields []string
}

var defaultListFields = []string{"_id", "name", "productType", "status", "deprecated"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List products",
		Long:    "List products on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List products with default pagination
  incloud product list

  # Paginate
  incloud product list --page 2 --limit 50

  # Filter by name (LIKE search)
  incloud product list --name IR615

  # Filter by product type
  incloud product list --type router

  # Filter by status
  incloud product list --status PUBLISHED

  # Table output with selected fields
  incloud product list -o table -f name -f productType -f status`,
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
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.Type != "" {
				q.Set("productType", opts.Type)
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

			body, err := client.Get("/api/v1/products", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort order (e.g. \"createdAt,desc\")")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by product type")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (INDEVELOPMENT|PUBLISHED)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
