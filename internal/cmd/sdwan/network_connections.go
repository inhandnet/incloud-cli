package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultConnectionFields = []string{"_id", "source", "target", "status"}

type networkConnectionsOptions struct {
	cmdutil.ListFlags
	Name string
}

func newCmdNetworkConnections(f *factory.Factory) *cobra.Command {
	opts := &networkConnectionsOptions{}

	cmd := &cobra.Command{
		Use:   "connections <networkId>",
		Short: "List connections in an SD-WAN network",
		Example: `  # List all connections
  incloud sdwan network connections <id>

  # Filter by device name
  incloud sdwan network connections <id> --name ER805`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultConnectionFields)
			if opts.Name != "" {
				q.Set("deviceName", opts.Name)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/connections", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")

	return cmd
}
