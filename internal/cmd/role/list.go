package role

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ListOptions struct {
	Page   int
	Limit  int
	Sort   string
	App    string
	Fields []string
}

var defaultListFields = []string{"_id", "name", "description", "builtInRole", "subOrgVisible"}

func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List roles",
		Long:    "List roles on the InCloud platform with optional filtering and pagination.",
		Aliases: []string{"ls"},
		Example: `  # List roles
  incloud role list

  # Filter by application
  incloud role list --app portal

  # Table output with selected fields
  incloud role list -o table -f _id -f name -f builtInRole`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultListFields)
			if opts.App != "" {
				q.Set("app", opts.App)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/roles", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.App, "app", "", "Filter by application (e.g. portal, console)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
