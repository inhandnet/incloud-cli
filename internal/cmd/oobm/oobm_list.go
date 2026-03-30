package oobm

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmListOptions struct {
	Name     string
	DeviceID string
	cmdutil.ListFlags
}

var defaultOobmListFields = []string{"_id", "name", "deviceId", "clientIp", "services", "idleTime", "connTime", "createdAt"}

func NewCmdOobmList(f *factory.Factory) *cobra.Command {
	opts := &OobmListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List OOBM resources",
		Aliases: []string{"ls"},
		Example: `  # List OOBM resources
  incloud oobm list

  # Filter by device
  incloud oobm list --device-id 507f1f77bcf86cd799439011

  # Filter by name
  incloud oobm list --name "Router SSH"

  # Paginate
  incloud oobm list --page 2 --limit 50

  # Table with selected fields
  incloud oobm list -o table -f _id -f name -f clientIp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultOobmListFields)
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/oobm/resources", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID (use 'incloud device list' to find IDs)")
	opts.ListFlags.Register(cmd)

	return cmd
}
