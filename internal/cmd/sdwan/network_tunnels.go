package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultTunnelFields = []string{"_id", "source.deviceName", "target.deviceName", "source.interfaceName", "target.interfaceName", "status", "stateUpdatedAt"}

type networkTunnelsOptions struct {
	cmdutil.ListFlags
	Name     string
	DeviceID string
}

func newCmdNetworkTunnels(f *factory.Factory) *cobra.Command {
	opts := &networkTunnelsOptions{}

	cmd := &cobra.Command{
		Use:   "tunnels <networkId>",
		Short: "List tunnels in an SD-WAN network",
		Example: `  # List all tunnels
  incloud sdwan network tunnels <id>

  # Filter by device name
  incloud sdwan network tunnels <id> --name ER805

  # Filter by device ID
  incloud sdwan network tunnels <id> --device-id <deviceId>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultTunnelFields)
			if opts.Name != "" {
				q.Set("deviceName", opts.Name)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/tunnels", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID")

	return cmd
}
