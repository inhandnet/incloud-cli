package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page    int
	Limit   int
	Sort    string
	Name    string
	Type    string
	Status  string
	Columns []string
}

var defaultListColumns = []string{"_id", "name", "productType", "status", "deprecated"}

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
  incloud product list --page 1 --limit 50

  # Filter by name (LIKE search)
  incloud product list --name IR615

  # Filter by product type
  incloud product list --type router

  # Filter by status
  incloud product list --status PUBLISHED

  # Table output with selected columns
  incloud product list -o table -c name -c productType -c status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			u, err := url.Parse(ctx.Host + "/api/v1/products")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			q.Set("page", strconv.Itoa(opts.Page))
			q.Set("size", strconv.Itoa(opts.Limit))
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
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "table":
				columns := opts.Columns
				if len(columns) == 0 && f.IO.IsStdoutTTY() {
					columns = defaultListColumns
				}
				if err := iostreams.FormatTable(body, f.IO, columns); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 0, "Page number (default 0)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page (default 20)")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", "Sort order (e.g. \"createdAt,desc\")")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by product type")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (INDEVELOPMENT|PUBLISHED)")
	cmd.Flags().StringArrayVarP(&opts.Columns, "column", "c", nil, "Columns to show in table output")

	return cmd
}
