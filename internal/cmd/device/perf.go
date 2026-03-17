package device

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type perfOptions struct {
	After   string
	Before  string
	Refresh bool
	Fields  []string
}

var defaultPerfCurrentFields = []string{"cpu.usage", "memory.free", "memory.total", "disk.free", "disk.total", "updatedAt"}
var defaultPerfSeriesFields = []string{"time", "cpuUsage", "cpu0", "cpu1", "cpu2", "cpu3", "memoryUsage"}

func NewCmdPerf(f *factory.Factory) *cobra.Command {
	opts := &perfOptions{}

	cmd := &cobra.Command{
		Use:   "perf <device-id>",
		Short: "Show device performance (CPU, memory, disk)",
		Long: `Show device performance metrics.

By default, displays the current performance snapshot (CPU usage, memory, disk).
With --after/--before, displays historical performance time series.
With --refresh, triggers a real-time collection from the device (online only).`,
		Example: `  # Current performance snapshot
  incloud device perf 507f1f77bcf86cd799439011

  # Real-time collection (device must be online)
  incloud device perf 507f1f77bcf86cd799439011 --refresh

  # Historical time series
  incloud device perf 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00 --before 2024-01-02T00:00:00

  # Table output with selected fields
  incloud device perf 507f1f77bcf86cd799439011 -o table -f cpu.usage -f memory.free -f memory.total`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			if opts.After != "" || opts.Before != "" {
				return runPerfSeries(f, cmd, deviceID, opts)
			}
			return runPerfCurrent(f, cmd, deviceID, opts)
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().BoolVar(&opts.Refresh, "refresh", false, "Trigger real-time collection (device must be online)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}

func runPerfCurrent(f *factory.Factory, cmd *cobra.Command, deviceID string, opts *perfOptions) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	actx, err := cfg.ActiveContext()
	if err != nil {
		return err
	}
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Trigger real-time collection first, then always GET cached data
	if opts.Refresh {
		refreshURL := actx.Host + "/api/v1/devices/" + deviceID + "/performances/refresh"
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, refreshURL, http.NoBody)
		if err != nil {
			return fmt.Errorf("building refresh request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("refresh request failed: %w", err)
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			fmt.Fprintln(f.IO.ErrOut, "Warning: refresh failed (device may be offline), showing cached data")
		}
	}

	endpoint := actx.Host + "/api/v1/devices/" + deviceID + "/performances"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, http.NoBody)
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
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultPerfCurrentFields
	}
	return iostreams.FormatOutput(body, f.IO, output, fields)
}

func runPerfSeries(f *factory.Factory, cmd *cobra.Command, deviceID string, opts *perfOptions) error {
	if opts.After == "" || opts.Before == "" {
		return fmt.Errorf("both --after and --before are required for historical data")
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}
	actx, err := cfg.ActiveContext()
	if err != nil {
		return err
	}
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	u, err := url.Parse(actx.Host + "/api/v1/devices/" + deviceID + "/performance")
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("after", opts.After)
	q.Set("before", opts.Before)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
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
	fields := opts.Fields
	if len(fields) == 0 {
		fields = defaultPerfSeriesFields
	}
	return iostreams.FormatOutput(body, f.IO, output, fields, iostreams.WithTransform(iostreams.FlattenSeries))
}
