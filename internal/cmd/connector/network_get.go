package connector

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type networkGetOptions struct {
	Fields []string
	Expand string
}

func newCmdNetworkGet(f *factory.Factory) *cobra.Command {
	opts := &networkGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get connector network details",
		Example: `  # Get network details
  incloud connector network get 66827b3ccfb1842140f4222f

  # With live connected counts
  incloud connector network get 66827b3ccfb1842140f4222f --expand counts`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if opts.Expand != "" {
				q.Set("expand", opts.Expand)
			}
			if len(opts.Fields) > 0 {
				q.Set("fields", strings.Join(opts.Fields, ","))
			}

			body, err := client.Get("/api/v1/connectors/"+args[0], q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", `Expand related data (e.g. "counts" for live connected counts)`)

	return cmd
}
