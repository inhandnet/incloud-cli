package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultNetworkFields = []string{"_id", "name", "type", "tunnelCreationMode", "totalDevices", "hubs", "spokes", "createdAt"}

type networkListOptions struct {
	Page   int
	Limit  int
	Sort   string
	Name   string
	Fields []string
}

func newCmdNetworkList(f *factory.Factory) *cobra.Command {
	opts := &networkListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List SD-WAN networks",
		Example: `  # List all networks
  incloud sdwan network list

  # Filter by name
  incloud sdwan network list --name my-sdwan

  # Custom fields
  incloud sdwan network list -f _id -f name -f totalDevices`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultNetworkFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get(apiBase+"/networks", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
