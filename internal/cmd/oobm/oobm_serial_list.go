package oobm

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmSerialListOptions struct {
	Name     string
	DeviceID string
	cmdutil.ListFlags
}

var defaultSerialListFields = []string{
	"_id", "name", "deviceId", "speed", "dataBits", "parity", "usage", "connected", "url", "createdAt",
}

func NewCmdOobmSerialList(f *factory.Factory) *cobra.Command {
	opts := &OobmSerialListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OOBM serial port configurations",
		Aliases: []string{"ls"},
		Example: `  # List serial port configurations
  incloud oobm serial list

  # Filter by device
  incloud oobm serial list --device-id 507f1f77bcf86cd799439011

  # Table with selected fields
  incloud oobm serial list -o table -f _id -f name -f speed -f connected`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultSerialListFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/oobm/serials", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID (use 'incloud device list' to find IDs)")
	opts.Register(cmd)

	return cmd
}
