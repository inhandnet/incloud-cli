package user

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultMeFields = []string{"_id", "username", "name", "email", "oid", "locale", "blocked"}

type MeOptions struct {
	Fields []string
	Expand string
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
  incloud user me -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var q url.Values
			if len(opts.Fields) > 0 || opts.Expand != "" {
				q = url.Values{}
			}
			if len(opts.Fields) > 0 {
				q.Set("fields", strings.Join(opts.Fields, ","))
			}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}

			body, err := client.Get("/api/v1/users/me", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultMeFields
			}
			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", "Expand related objects (e.g. roles)")

	return cmd
}
