package oobm

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmLogsOptions struct {
	Type       string
	Protocols  []string
	BusinessID string
	Page       int
	Limit      int
	Sort       string
	Fields     []string
}

var defaultOobmLogsFields = []string{
	"_id", "proto", "clientIp", "port", "type", "status",
	"createdAt", "endedAt", "sentBytes", "recvBytes",
}

func NewCmdOobmLogs(f *factory.Factory) *cobra.Command {
	opts := &OobmLogsOptions{}

	cmd := &cobra.Command{
		Use:   "logs <device-id>",
		Short: "List OOBM tunnel connection logs",
		Long:  "List tunnel connection logs for a device, with optional filtering by type, protocol, and business resource.",
		Example: `  # List logs for a device
  incloud oobm logs 507f1f77bcf86cd799439011

  # Filter by protocol
  incloud oobm logs 507f1f77bcf86cd799439011 --protocol ssh --protocol tcp

  # Filter by type and business ID
  incloud oobm logs 507f1f77bcf86cd799439011 --type oobm --business-id abc123

  # Paginate and sort
  incloud oobm logs 507f1f77bcf86cd799439011 --page 2 --limit 50 --sort "createdAt,desc"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := make(url.Values)
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))

			if opts.Type != "" {
				q.Set("type", opts.Type)
			}
			if len(opts.Protocols) > 0 {
				q.Set("protocols", strings.Join(opts.Protocols, ","))
			}
			if opts.BusinessID != "" {
				q.Set("businessId", opts.BusinessID)
			}
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultOobmLogsFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/ngrok/devices/"+deviceID+"/logs", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringVar(&opts.Type, "type", "", "Tunnel type filter")
	cmd.Flags().StringSliceVar(&opts.Protocols, "protocol", nil, "Protocol filter (can be repeated)")
	cmd.Flags().StringVar(&opts.BusinessID, "business-id", "", "Business resource ID filter")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
