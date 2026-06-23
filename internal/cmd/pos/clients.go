package pos

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type clientsOptions struct {
	cmdutil.ListFlags
	Level    string
	DeviceID string
	OID      string
}

func newCmdClients(f *factory.Factory) *cobra.Command {
	opts := &clientsOptions{}

	cmd := &cobra.Command{
		Use:   "clients",
		Short: "List POS-marked clients across all devices",
		Long:  "List clients marked with a POS priority level (priority/bypass) across all your devices.",
		Example: `  # List all marked clients (priority + bypass)
  incloud pos clients

  # Only prioritized clients
  incloud pos clients --level priority

  # Filter by device, expand device and org info
  incloud pos clients --device 507f1f77bcf86cd799439011 --expand device,org`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			if opts.Level != "" {
				q.Set("level", opts.Level)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}
			if opts.OID != "" {
				q.Set("oid", opts.OID)
			}

			body, err := client.Get("/api/v1/pos-ready/clients", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Level, "level", "", "Filter by level (priority/bypass)")
	cmd.Flags().StringVar(&opts.DeviceID, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.OID, "oid", "", "Filter by organization ID")
	opts.Register(cmd)
	opts.RegisterExpand(cmd, "device", "org")

	return cmd
}
