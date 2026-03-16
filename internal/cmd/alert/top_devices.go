package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type TopDevicesOptions struct {
	After   string
	Before  string
	Group   []string
	N       int
	Columns []string
}

var defaultTopDevicesColumns = []string{"deviceId", "deviceName", "serialNumber", "value"}

func NewCmdTopDevices(f *factory.Factory) *cobra.Command {
	opts := &TopDevicesOptions{}

	cmd := &cobra.Command{
		Use:     "devices",
		Short:   "Top devices by alert count",
		Long:    "Show top-K devices ranked by alert count within a time range.",
		Aliases: []string{"device"},
		Example: `  # Top 10 devices with most alerts (last 30 days)
  incloud alert top devices

  # Top 5 devices
  incloud alert top devices --n 5

  # Filter by time range
  incloud alert top devices --after 2024-01-01T00:00:00 --before 2024-01-31T23:59:59

  # Filter by device group
  incloud alert top devices --group 507f1f77bcf86cd799439011

  # Table output with selected columns
  incloud alert top devices -o table -c deviceName -c value`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/alert/top-alert-devices")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			for _, g := range opts.Group {
				q.Add("devicegroupId", g)
			}
			q.Set("n", strconv.Itoa(opts.N))
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "table":
				columns := opts.Columns
				if len(columns) == 0 && f.IO.IsStdoutTTY() {
					columns = defaultTopDevicesColumns
				}
				if err := iostreams.FormatTable(body, f.IO, columns); err != nil {
					return err
				}
			case "yaml":
				s, err := iostreams.FormatYAML(body)
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, s)
			default:
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2024-01-31T23:59:59)")
	cmd.Flags().StringArrayVar(&opts.Group, "group", nil, "Filter by device group ID (can be repeated)")
	cmd.Flags().IntVar(&opts.N, "n", 10, "Number of top devices to return (default 10)")
	cmd.Flags().StringArrayVarP(&opts.Columns, "column", "c", nil, "Columns to show in table output")

	return cmd
}
