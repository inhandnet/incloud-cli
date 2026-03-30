package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultNetworkFields = []string{"_id", "name", "subnet", "counts.deviceCount", "counts.accountCount", "counts.endpointsCount", "createdAt"}

type networkListOptions struct {
	cmdutil.ListOpts
	Name string
}

func newCmdNetworkList(f *factory.Factory) *cobra.Command {
	opts := &networkListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List connector networks",
		Example: `  # List all connector networks
  incloud connector network list

  # Filter by name
  incloud connector network list --name my-vpn

  # Custom fields
  incloud connector network list -f _id -f name -f subnet`,
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

			body, err := client.Get("/api/v1/connectors", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmdutil.RegisterListFlags(cmd, &opts.ListOpts)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")

	return cmd
}
