package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdNetworkConnectionTunnels(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection-tunnels <networkId> <connectionId>",
		Short: "List tunnels for a specific connection",
		Example: `  # List tunnels for a connection
  incloud sdwan network connection-tunnels <networkId> <connectionId>`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/connections/"+args[1]+"/tunnels", nil)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body, nil)
		},
	}

	return cmd
}
