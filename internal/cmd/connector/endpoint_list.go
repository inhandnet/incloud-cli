package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultEndpointFields = []string{"_id", "name", "lanIp", "vip", "deviceName", "connected", "createdAt"}

type endpointListOptions struct {
	cmdutil.ListFlags
	Name     string
	LanIP    string
	DeviceID string
	Search   string
}

func newCmdEndpointList(f *factory.Factory) *cobra.Command {
	opts := &endpointListOptions{}

	cmd := &cobra.Command{
		Use:     "list <network-id>",
		Aliases: []string{"ls"},
		Short:   "List endpoints in a connector network",
		Example: `  # List all endpoints in a network
  incloud connector endpoint list 66827b3ccfb1842140f4222f

  # Filter by name
  incloud connector endpoint list 66827b3ccfb1842140f4222f --name my-endpoint

  # Search by name or LAN IP
  incloud connector endpoint list 66827b3ccfb1842140f4222f -q 192.168

  # Custom fields
  incloud connector endpoint list 66827b3ccfb1842140f4222f -f _id -f name -f lanIp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultEndpointFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.LanIP != "" {
				q.Set("lanIp", opts.LanIP)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}
			if opts.Search != "" {
				q.Set("nameOrLanIp", opts.Search)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/connectors/"+networkID+"/endpoints", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.LanIP, "lan-ip", "", "Filter by LAN IP")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID (use 'incloud connector device list <network-id>' to find IDs)")
	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or LAN IP")

	return cmd
}
