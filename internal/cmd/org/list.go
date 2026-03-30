package org

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	cmdutil.ListFlags
	Name         string
	Email        string
	ContactEmail string
	Q            string
	Ancestor     string
	Depth        int
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

			q := cmdutil.NewQuery(cmd, defaultListFields)
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

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/orgs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")
	cmd.Flags().StringVar(&opts.Email, "email", "", "Filter by email (LIKE search)")
	cmd.Flags().StringVar(&opts.ContactEmail, "contact-email", "", "Filter by contact email (LIKE search)")
	cmd.Flags().StringVarP(&opts.Q, "search", "q", "", "General search query")
	cmd.Flags().StringVar(&opts.Ancestor, "ancestor", "", "Filter by ancestor organization ID (use 'incloud org list' or 'incloud org self' to find IDs)")
	cmd.Flags().IntVar(&opts.Depth, "depth", 0, "Organization tree depth (default: API returns depth 1)")
	opts.ListFlags.RegisterExpand(cmd)

	return cmd
}
