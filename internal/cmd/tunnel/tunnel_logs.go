package tunnel

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TunnelLogsOptions struct {
	Type       string
	Protocols  []string
	BusinessID string
	cmdutil.ListFlags
}

var defaultTunnelLogsFields = []string{
	"_id", "proto", "clientIp", "port", "type", "status",
	"createdAt", "endedAt", "sentBytes", "recvBytes",
}

func NewCmdTunnelLogs(f *factory.Factory) *cobra.Command {
	opts := &TunnelLogsOptions{}

	cmd := &cobra.Command{
		Use:   "logs <device-id>",
		Short: "List tunnel connection logs",
		Long:  "List tunnel connection logs for a device, with optional filtering by type and protocol.",
		Example: `  # List tunnel logs for a device
  incloud tunnel logs 507f1f77bcf86cd799439011

  # Filter by protocol
  incloud tunnel logs 507f1f77bcf86cd799439011 --protocol local_web

  # Paginate
  incloud tunnel logs 507f1f77bcf86cd799439011 --page 2 --limit 50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, defaultTunnelLogsFields)
			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if len(opts.Protocols) > 0 {
				q.Set("protocols", strings.Join(opts.Protocols, ","))
			}
			if opts.BusinessID != "" {
				q.Set("businessId", opts.BusinessID)
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/ngrok/devices/"+deviceID+"/logs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "local", "Tunnel type filter")
	cmd.Flags().StringSliceVar(&opts.Protocols, "protocol", nil, "Protocol filter: local_web, local_cli (can be repeated)")
	cmd.Flags().StringVar(&opts.BusinessID, "business-id", "", "Business resource ID filter")
	opts.ListFlags.Register(cmd)

	return cmd
}
