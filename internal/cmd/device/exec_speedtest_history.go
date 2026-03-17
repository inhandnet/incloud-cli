package device

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

var defaultSpeedtestHistoryFields = []string{"_id", "interface", "download", "upload", "idleLatency", "jitter", "loss", "serverNode", "success", "createdAt"}

func NewCmdExecSpeedtestHistory(f *factory.Factory) *cobra.Command {
	var (
		after  string
		before string
		page   int
		limit  int
		fields []string
	)

	cmd := &cobra.Command{
		Use:   "speedtest-history <device-id>",
		Short: "View speed test history for a device",
		Example: `  # View recent speed test results
  incloud device exec speedtest-history 507f1f77bcf86cd799439011

  # Filter by date range
  incloud device exec speedtest-history 507f1f77bcf86cd799439011 --after 2024-01-01 --before 2024-02-01`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

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

			u, err := url.Parse(actx.Host + "/api/v1/devices/" + deviceID + "/diagnosis/speed-test-histories")
			if err != nil {
				return err
			}
			q := u.Query()
			q.Set("page", strconv.Itoa(page-1))
			q.Set("size", strconv.Itoa(limit))
			if after != "" {
				q.Set("from", after)
			}
			if before != "" {
				q.Set("to", before)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			cols := defaultSpeedtestHistoryFields
			if len(fields) > 0 {
				cols = fields
			}
			if err := formatOutput(cmd, f.IO, respBody, cols); err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&before, "before", "", "End date (ISO 8601)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number (1-based)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Results per page")
	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to display (comma-separated)")

	return cmd
}
