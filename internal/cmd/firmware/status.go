package firmware

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type StatusOptions struct {
	Page    int
	Limit   int
	Sort    string
	Device  string
	Product string
	Module  string
	Status  string
	Version string
	Expand  []string
	Fields  []string
}

var defaultStatusFields = []string{
	"deviceId", "product", "module", "currentVersion",
	"status", "pendingVersion", "latestVersion", "statusUpdatedAt",
}

func NewCmdStatus(f *factory.Factory) *cobra.Command {
	opts := &StatusOptions{}

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "List device firmware upgrade status",
		Long:    "List device firmware and OTA module upgrade status with optional filtering.",
		Aliases: []string{"st"},
		Example: `  # List all devices' firmware status
  incloud firmware status

  # Filter by product
  incloud firmware status --product ER805

  # Filter by upgrade status
  incloud firmware status --status queued

  # Show all OTA modules for a specific device
  incloud firmware status --device 6989ad34a7455f3f0bf9dce2

  # Show a specific module for a device
  incloud firmware status --device 6989ad34a7455f3f0bf9dce2 --module modem

  # Expand device info
  incloud firmware status --expand device`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(opts.Page-1))
			q.Set("limit", strconv.Itoa(opts.Limit))
			if opts.Sort != "" {
				q.Set("sort", opts.Sort)
			}
			if opts.Module != "" {
				q.Set("module", opts.Module)
			}
			if opts.Status != "" {
				q.Set("status", opts.Status)
			}
			if opts.Version != "" {
				q.Set("currentVersion", opts.Version)
			}
			if len(opts.Expand) > 0 {
				q.Set("expand", strings.Join(opts.Expand, ","))
			}

			output, _ := cmd.Flags().GetString("output")
			fields := opts.Fields
			if len(fields) == 0 && output == "table" {
				fields = defaultStatusFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
			}

			// Choose endpoint based on --device flag
			var path string
			if opts.Device != "" {
				path = "/api/v1/devices/" + url.PathEscape(opts.Device) + "/ota/modules"
			} else {
				path = "/api/v1/device/firmwares"
				if opts.Product != "" {
					q.Set("product", opts.Product)
				}
			}

			body, err := client.Get(path, q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.Sort, "sort", "", `Sort order (e.g. "createdAt,desc")`)
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID (shows all OTA modules for the device)")
	cmd.Flags().StringVar(&opts.Product, "product", "", "Filter by product name")
	cmd.Flags().StringVar(&opts.Module, "module", "", "Filter by module name")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (up_to_date|new_firmware_available|queued|in_progress)")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Filter by current firmware version")
	cmd.Flags().StringSliceVar(&opts.Expand, "expand", nil, "Expand related objects (e.g. device)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
