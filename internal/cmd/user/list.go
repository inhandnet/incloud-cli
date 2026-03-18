package user

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
	Email  string
	Name   string
	Q      string
	Type   string
	Expand string
	Fields []string
}

var defaultListFields = []string{"_id", "username", "name", "email", "blocked", "lastLogin"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List users",
		Long:    "List users on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List users with default pagination
  incloud user list

  # Paginate
  incloud user list --page 2 --limit 50

  # Search by email
  incloud user list --email example.com

  # Search by name
  incloud user list --name john

  # General search
  incloud user list --search admin

  # Filter by type
  incloud user list --type INTERNAL

  # Expand roles
  incloud user list --expand roles

  # Table output with selected fields
  incloud user list -o table -f username -f email -f name`,
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
			if opts.Email != "" {
				q.Set("email", opts.Email)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.Q != "" {
				q.Set("q", opts.Q)
			}
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/users", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Email, "email", "", "Filter by email (LIKE search)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVarP(&opts.Q, "search", "q", "", "General search query")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by user type (INTERNAL=org members, EXTERNAL=collaborators)")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", `Expand related resources (e.g. "roles")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
