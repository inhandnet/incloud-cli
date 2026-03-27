package org

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page         int
	Limit        int
	Sort         string
	Name         string
	Email        string
	ContactEmail string
	Q            string
	Ancestor     string
	Depth        int
	Expand       string
	Fields       []string
}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List organizations",
		Long:    "List organizations on the InCloud platform with optional filtering and pagination.",
		Example: `  # List organizations
  incloud org list

  # Search by name
  incloud org list --search "Acme"

  # Filter by ancestor
  incloud org list --ancestor 61259f8f4be3e571fcfa4d75

  # Paginate
  incloud org list --page 2 --limit 50

  # Select fields
  incloud org list -f _id -f name -f email`,
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
			if opts.Email != "" {
				q.Set("email", opts.Email)
			}
			if opts.ContactEmail != "" {
				q.Set("contactEmail", opts.ContactEmail)
			}
			if opts.Q != "" {
				q.Set("q", opts.Q)
			}
			if opts.Ancestor != "" {
				q.Set("ancestor", opts.Ancestor)
			}
			if opts.Depth != 0 {
				q.Set("depth", strconv.Itoa(opts.Depth))
			}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/orgs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Filter by email (LIKE search)")
	cmd.Flags().StringVar(&opts.ContactEmail, "contact-email", "", "Filter by contact email (LIKE search)")
	cmd.Flags().StringVarP(&opts.Q, "search", "q", "", "General search query")
	cmd.Flags().StringVar(&opts.Ancestor, "ancestor", "", "Filter by ancestor organization ID (use 'incloud org list' or 'incloud org self' to find IDs)")
	cmd.Flags().IntVar(&opts.Depth, "depth", 0, "Organization tree depth (default: API returns depth 1)")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", `Expand related resources (e.g. "parent")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
