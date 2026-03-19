package sdwan

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultTunnelFields = []string{"_id", "source.deviceName", "target.deviceName", "source.interfaceName", "target.interfaceName", "status", "stateUpdatedAt"}

type networkTunnelsOptions struct {
	Page     int
	Limit    int
	Sort     string
	Name     string
	DeviceID string
	Fields   []string
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

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Name != "" {
				q.Set("deviceName", opts.Name)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultTunnelFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/tunnels", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
