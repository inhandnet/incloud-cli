package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

type networkGetOptions struct {
	Fields []string
	Expand []string
}

func newCmdNetworkGet(f *factory.Factory) *cobra.Command {
	opts := &networkGetOptions{}

	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get SD-WAN network details",
		Example: `  # Get network details (includes tunnel counts by default)
  incloud sdwan network get <id>

  # Without tunnel counts
  incloud sdwan network get <id> --expand ""`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)

			body, err := client.Get(apiBase+"/networks/"+args[0], q)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", []string{"tunnels"}, `Expand related data (e.g. "tunnels" for tunnel counts)`)

	return cmd
}
