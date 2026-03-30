package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

var defaultDeviceFields = []string{"_id", "deviceName", "serialNumber", "role", "product", "online", "createdAt"}

type devicesOptions struct {
	cmdutil.ListFlags
	Role         string
	Name         string
	SerialNumber string
	Product      []string
}

func newCmdDevices(f *factory.Factory) *cobra.Command {
	opts := &devicesOptions{}

	cmd := &cobra.Command{
		Use:   "devices <networkId>",
		Short: "List devices in an SD-WAN network",
		Example: `  # List all devices in a network
  incloud sdwan devices <networkId>

  # Filter by role
  incloud sdwan devices <networkId> --role hub

  # Filter by name
  incloud sdwan devices <networkId> --name ER805

  # Filter by product
  incloud sdwan devices <networkId> --product ER805 --product MR805`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultDeviceFields)
			if opts.Role != "" {
				q.Set("role", opts.Role)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.SerialNumber != "" {
				q.Set("serialNumber", opts.SerialNumber)
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get(apiBase+"/networks/"+args[0]+"/devices", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.ListFlags.Register(cmd)
	cmd.Flags().StringVar(&opts.Role, "role", "", "Filter by role: hub or spoke")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name")
	cmd.Flags().StringVar(&opts.SerialNumber, "serial-number", "", "Filter by serial number")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product model (repeatable)")

	return cmd
}
