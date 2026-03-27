package alert

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <alert-id>",
		Short: "Get alert details",
		Long:  "Get detailed information about a specific alert by its ID.",
		Example: `  # Get alert details (colorized JSON in TTY)
  incloud alert get 507f1f77bcf86cd799439011

  # Only specific fields
  incloud alert get 507f1f77bcf86cd799439011 -f type -f status -f entityName

  # Table output (KEY/VALUE pairs)
  incloud alert get 507f1f77bcf86cd799439011 -o table

  # YAML output
  incloud alert get 507f1f77bcf86cd799439011 -o yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alertID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var q url.Values
			if len(opts.Fields) > 0 {
				q = make(url.Values)
				q.Set("fields", strings.Join(opts.Fields, ","))
			}

			body, err := client.Get("/api/v1/alerts/"+alertID, q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
