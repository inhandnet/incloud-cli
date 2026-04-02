package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultDeviceFields = []string{"_id", "serialNumber", "name", "vip", "subnet", "connected", "createdAt"}

type deviceListOptions struct {
	cmdutil.ListFlags
	Name      string
	SN        string
	Connected string
	Search    string
}

func newCmdDeviceList(f *factory.Factory) *cobra.Command {
	opts := &deviceListOptions{}

	cmd := &cobra.Command{
		Use:     "list <network-id>",
		Aliases: []string{"ls"},
		Short:   "List devices in a connector network",
		Example: `  # List devices in a network
  incloud connector device list 66827b3ccfb1842140f4222f

  # Filter connected devices
  incloud connector device list 66827b3ccfb1842140f4222f --connected true`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultDeviceFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.SN != "" {
				q.Set("serialNumber", opts.SN)
			}
			if opts.Connected != "" {
				q.Set("connected", opts.Connected)
			}
			if opts.Search != "" {
				q.Set("nameOrSn", opts.Search)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/connectors/"+networkID+"/devices", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")
	cmd.Flags().StringVar(&opts.SN, "sn", "", "Filter by serial number")
	cmd.Flags().StringVar(&opts.Connected, "connected", "", "Filter by connected status (true/false)")
	cmd.Flags().StringVarP(&opts.Search, "search", "q", "", "Search by name or serial number")

	return cmd
}
