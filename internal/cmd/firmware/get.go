package firmware

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
		Use:   "get <id>",
		Short: "Get firmware details",
		Long:  "Get detailed information about a specific firmware by its ID.",
		Example: `  # Get firmware by ID
  incloud firmware get 507f1f77bcf86cd799439011

  # Table output
  incloud firmware get 507f1f77bcf86cd799439011 -o table

  # Select fields
  incloud firmware get 507f1f77bcf86cd799439011 -f product -f version -f status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var q url.Values
			if len(opts.Fields) > 0 {
				q = url.Values{}
				q.Set("fields", strings.Join(opts.Fields, ","))
			}

			body, err := client.Get("/api/v1/firmwares/"+url.PathEscape(id), q)
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
