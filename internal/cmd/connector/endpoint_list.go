package connector

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultEndpointFields = []string{"_id", "name", "lanIp", "vip", "deviceName", "connected", "createdAt"}

type endpointListOptions struct {
	Page     int
	Limit    int
	Sort     string
	Name     string
	LanIP    string
	DeviceID string
	Search   string
	Fields   []string
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

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
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
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultEndpointFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/connectors/"+networkID+"/endpoints", q)
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
	cmd.Flags().StringVar(&opts.LanIP, "lan-ip", "", "Filter by LAN IP")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID (use 'incloud connector device list <network-id>' to find IDs)")
	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or LAN IP")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
