package org

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type GetOptions struct {
	Fields []string
	Expand string
}

func NewCmdGet(f *factory.Factory) *cobra.Command {
	opts := &GetOptions{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get organization details",
		Example: `  # Get organization by ID
  incloud org get 61259f8f4be3e571fcfa4d75

  # With specific fields
  incloud org get 61259f8f4be3e571fcfa4d75 -f name -f email -f userCount

  # YAML output
  incloud org get 61259f8f4be3e571fcfa4d75 -o yaml`,
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

			body, err := client.Get("/api/v1/orgs/"+args[0], q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, opts.Fields)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringVar(&opts.Expand, "expand", "", `Expand related resources (e.g. "parent")`)

	return cmd
}
