package device

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type ExportOptions struct {
	Query        string
	Online       string
	Product      []string
	Group        []string
	Name         string
	SerialNumber string
	Firmware     string
	ConfigStatus []string
	IP           string
	MAC          string
	File         string
}

func NewCmdExport(f *factory.Factory) *cobra.Command {
	opts := &ExportOptions{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export devices to CSV",
		Long:  "Export devices as a server-generated CSV file. Supports filtering by device attributes.",
		Example: `  # Export all devices to stdout
  incloud device export

  # Export to a file
  incloud device export --file devices.csv

  # Export online devices only
  incloud device export --online true --file online.csv

  # Filter by product and search
  incloud device export --product IR915L -q "router"

  # Export devices with specific firmware
  incloud device export --firmware V1.0.0 --file fw100.csv

  # Export devices with pending config
  incloud device export --config-status PENDING

  # Pipe to other commands
  incloud device export | head -20`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if opts.Query != "" {
				q.Set("q", opts.Query)
			}
			if opts.Online != "" {
				q.Set("online", opts.Online)
			}
			if opts.Name != "" {
				q.Set("name", opts.Name)
			}
			if opts.SerialNumber != "" {
				q.Set("serialNumber", opts.SerialNumber)
			}
			if opts.Firmware != "" {
				q.Set("firmware", opts.Firmware)
			}
			if opts.IP != "" {
				q.Set("ip", opts.IP)
			}
			if opts.MAC != "" {
				q.Set("mac", opts.MAC)
			}
			for _, p := range opts.Product {
				q.Add("product", p)
			}
			for _, g := range opts.Group {
				q.Add("devicegroupId", g)
			}
			for _, s := range opts.ConfigStatus {
				q.Add("configStatus", s)
			}

			body, err := client.Get("/api/v1/devices/export", q)
			if err != nil {
				return err
			}

			if opts.File != "" {
				if err := os.WriteFile(opts.File, body, 0o600); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				fmt.Fprintf(f.IO.Out, "Exported to %s (%d bytes)\n", opts.File, len(body))
				return nil
			}

			_, err = f.IO.Out.Write(body)
			return err
		},
	}

	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by name or serial number")
	cmd.Flags().StringVar(&opts.Online, "online", "", "Filter by online status (true/false)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter by device name (fuzzy match)")
	cmd.Flags().StringVar(&opts.SerialNumber, "serial-number", "", "Filter by serial number (fuzzy match)")
	cmd.Flags().StringVar(&opts.Firmware, "firmware", "", "Filter by firmware version (fuzzy match)")
	cmd.Flags().StringArrayVar(&opts.ConfigStatus, "config-status", nil, "Filter by config status: SYNCED/PENDING/SUSPENDED/ERROR/NONE (can be repeated)")
	cmd.Flags().StringVar(&opts.IP, "ip", "", "Filter by IP address (fuzzy match)")
	cmd.Flags().StringVar(&opts.MAC, "mac", "", "Filter by MAC address (fuzzy match)")
	cmd.Flags().StringArrayVar(&opts.Product, "product", nil, "Filter by product (can be repeated)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().StringVar(&opts.File, "file", "", "Write output to file instead of stdout")

	return cmd
}
