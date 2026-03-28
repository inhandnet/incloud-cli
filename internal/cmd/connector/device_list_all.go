package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type deviceListAllOptions struct {
	Page      int
	Limit     int
	Sort      string
	NetworkID string
	Connected string
	Search    string
	Fields    []string
}

func newCmdDeviceListAll(f *factory.Factory) *cobra.Command {
	opts := &deviceListAllOptions{}

	cmd := &cobra.Command{
		Use:   "list-all",
		Short: "List all connector devices across networks",
		Example: `  # List all connector devices
  incloud connector device list-all

  # Filter by network
  incloud connector device list-all --network 66827b3ccfb1842140f4222f`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultDeviceFields)
			if opts.NetworkID != "" {
				q.Set("networkId", opts.NetworkID)
			}
			if opts.Connected != "" {
				q.Set("connected", opts.Connected)
			}
			if opts.Search != "" {
				q.Set("nameOrSn", opts.Search)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/connectors/devices", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.NetworkID, "network", "", "Filter by network ID (use 'incloud connector network list' to find IDs)")
	cmd.Flags().StringVar(&opts.Connected, "connected", "", "Filter by connected status (true/false)")
	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or serial number")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
