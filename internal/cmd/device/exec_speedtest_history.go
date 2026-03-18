package device

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
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

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("page", strconv.Itoa(page-1))
			q.Set("size", strconv.Itoa(limit))
			if after != "" {
				q.Set("from", after)
			}
			if before != "" {
				q.Set("to", before)
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/diagnosis/speed-test-histories", q)
			if err != nil {
				return err
			}

			cols := defaultSpeedtestHistoryFields
			if len(fields) > 0 {
				cols = fields
			}
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output, cols,
				iostreams.WithFormatters(iostreams.ColumnFormatters{
					"download": iostreams.FormatMbps,
					"upload":   iostreams.FormatMbps,
				}),
			)
		},
	}

	cmd.Flags().StringVar(&after, "after", "", "Start date (ISO 8601)")
	cmd.Flags().StringVar(&before, "before", "", "End date (ISO 8601)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number (1-based)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Results per page")
	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to display (comma-separated)")

	return cmd
}
