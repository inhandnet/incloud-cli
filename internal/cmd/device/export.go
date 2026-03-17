package device

import (
	"context"
	"fmt"
	"io"
	"net/http"
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
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			u, err := url.Parse(ctx.Host + "/api/v1/devices/export")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
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
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), "GET", u.String(), http.NoBody)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			w := f.IO.Out
			if opts.File != "" {
				file, err := os.Create(opts.File)
				if err != nil {
					return fmt.Errorf("creating file: %w", err)
				}
				defer func() { _ = file.Close() }()
				w = file
			}

			n, err := io.Copy(w, resp.Body)
			if err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			if opts.File != "" {
				fmt.Fprintf(f.IO.Out, "Exported to %s (%d bytes)\n", opts.File, n)
			}

			return nil
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
