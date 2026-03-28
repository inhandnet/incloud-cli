package user

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type MeOptions struct {
	Fields []string
	Expand []string
}

func NewCmdMe(f *factory.Factory) *cobra.Command {
	opts := &MeOptions{}

	cmd := &cobra.Command{
		Use:   "me",
		Short: "Show current user profile",
		Long:  "Show the profile of the currently authenticated user.",
		Example: `  # Show current user profile
  incloud user me

  # Table output with default fields
  incloud user me -o table

  # Only specific fields
  incloud user me -f username -f email -f locale

  # Expand roles
  incloud user me --expand roles

  # JSON output
  incloud user me -o json

  # Extract email with jq
  incloud user me --jq '.email'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)

			body, err := client.Get("/api/v1/users/me", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related objects (e.g. roles)")

	return cmd
}
