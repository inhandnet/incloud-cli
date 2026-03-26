package oobm

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type OobmListOptions struct {
	Name     string
	DeviceID string
	Page     int
	Limit    int
	Sort     string
	Fields   []string
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

			q := make(url.Values)
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.DeviceID != "" {
				q.Set("deviceId", opts.DeviceID)
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultOobmListFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			body, err := client.Get("/api/v1/oobm/resources", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output, fields)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by name")
	cmd.Flags().StringVar(&opts.DeviceID, "device-id", "", "Filter by device ID (use 'incloud device list' to find IDs)")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
