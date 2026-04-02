package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type clientListOptions struct {
	cmdutil.ListFlags
	Query  string
	Type   string
	Online string
	Device string
	MAC    string
	IP     string
	Asset  string
}

var defaultClientListFields = []string{"_id", "name", "mac", "ip", "type", "online", "deviceId", "ssid", "connectedAt"}

func newCmdClientList(f *factory.Factory) *cobra.Command {
	opts := &clientListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List connected clients",
		Long:    "List all clients (Wi-Fi/LAN devices) connected to your routers.",
		Aliases: []string{"ls"},
		Example: `  # List all clients
  incloud device client list

  # Filter by type
  incloud device client list --type wireless

  # Filter by online status
  incloud device client list --online true

  # Filter by device
  incloud device client list --device 507f1f77bcf86cd799439011

  # Search by name
  incloud device client list -q "desktop"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultClientListFields)
			if opts.Query != "" {
				q.Set("name", opts.Query)
			}
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if opts.Online != "" {
				q.Set("online", opts.Online)
			}
			if opts.Device != "" {
				q.Set("deviceId", opts.Device)
			}
			if opts.MAC != "" {
				q.Set("mac", opts.MAC)
			}
			if opts.IP != "" {
				q.Set("ip", opts.IP)
			}
			if opts.Asset != "" {
				q.Set("asset", opts.Asset)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/network/clients", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by client name")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by type (wireless/wired)")
	cmd.Flags().StringVar(&opts.Online, "online", "", "Filter by online status (true/false)")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.MAC, "mac", "", "Filter by MAC address")
	cmd.Flags().StringVar(&opts.IP, "ip", "", "Filter by IP address")
	cmd.Flags().StringVar(&opts.Asset, "asset", "", "Filter by asset status (true/false)")

	return cmd
}
