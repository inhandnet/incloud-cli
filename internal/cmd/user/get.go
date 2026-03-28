package user

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
	Expand []string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get user details",
		Long:  "Get detailed information about a specific user by ID.",
		Example: `  # Get user by ID
  incloud user get 69798301dfd35d106920c7b8

  # Only specific fields
  incloud user get 69798301dfd35d106920c7b8 -f name -f email -f username

  # Without expanding roles
  incloud user get 69798301dfd35d106920c7b8 --expand ""

  # Table output (KEY/VALUE pairs)
  incloud user get 69798301dfd35d106920c7b8 -o table

  # YAML output
  incloud user get 69798301dfd35d106920c7b8 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)

			body, err := client.Get("/api/v1/users/"+id, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", []string{"roles"}, "Related resources to expand")

	return cmd
}
