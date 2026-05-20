package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type clientListOptions struct {
	cmdutil.ListFlags
	DeviceID string
	Status   string
	Name     string
}

func newCmdClientList(f *factory.Factory) *cobra.Command {
	opts := &clientListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List touch clients",
		Long:    "List remote access clients with optional filtering by device, status, or name.",
		Aliases: []string{"ls"},
		Example: `  # List all clients
  incloud touch client list

  # Filter by device
  incloud touch client list --device-id 507f1f77bcf86cd799439011

  # Filter by connection status
  incloud touch client list --status CONNECTED

  # Filter by name
  incloud touch client list --name my-plc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}
			if opts.Status != "" {
				q.Set("touchConnectionStatus", opts.Status)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/touch/clients", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	opts.Register(cmd)
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by connection status (CONNECTED|DISCONNECTED|CONNECTING)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name (LIKE search)")

	return cmd
}
