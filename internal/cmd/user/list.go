package user

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Email  string
	Name   string
	Q      string
	Type   string
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

			q := cmdutil.NewQuery(cmd, defaultListFields)
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

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/users", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Email, "email", "", "Filter by email (LIKE search)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVarP(&opts.Q, "search", "q", "", "General search query")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by user type (INTERNAL=org members, EXTERNAL=collaborators)")
	opts.ListFlags.RegisterExpand(cmd, "roles", "org")

	return cmd
}
