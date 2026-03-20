package device

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type onlineOptions struct {
	Daily  bool
	Page   int
	Limit  int
	After  string
	Before string
	Fields []string
}

var defaultOnlineEventFields = []string{"timestamp", "eventType", "ipAddress", "disconnectReason"}

func NewCmdOnline(f *factory.Factory) *cobra.Command {
	opts := &onlineOptions{}

	cmd := &cobra.Command{
		Use:   "online <device-id>",
		Short: "Device online/offline history",
		Long: `View device online/offline history.

By default, shows individual connect/disconnect events (paginated).
Use --daily to show daily offline statistics instead (last 30 days).`,
		Example: `  # List online/offline events
  incloud device online 507f1f77bcf86cd799439011

  # Filter by time range
  incloud device online 507f1f77bcf86cd799439011 --after 2025-01-01T00:00:00 --before 2025-01-31T23:59:59

  # Daily offline statistics
  incloud device online 507f1f77bcf86cd799439011 --daily

  # Daily stats for a specific month
  incloud device online 507f1f77bcf86cd799439011 --daily --after 2025-03-01T00:00:00 --before 2025-03-31T00:00:00`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Daily {
				return runOnlineDaily(f, args[0], opts, cmd)
			}
			return runOnlineEvents(f, args[0], opts, cmd)
		},
	}

	cmd.Flags().BoolVar(&opts.Daily, "daily", false, "Show daily offline statistics instead of individual events")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number, starting from 1 (events mode only)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page (events mode only)")
	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2025-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2025-01-31T23:59:59)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}

func runOnlineEvents(f *factory.Factory, deviceID string, opts *onlineOptions, cmd *cobra.Command) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	q.Set("page", strconv.Itoa(opts.Page-1))
	q.Set("limit", strconv.Itoa(opts.Limit))
	if opts.After != "" {
		q.Set("from", opts.After)
	}
	if opts.Before != "" {
		q.Set("to", opts.Before)
	}

	output, _ := cmd.Flags().GetString("output")
	fields := opts.Fields
	if len(fields) == 0 && output == "table" {
		fields = defaultOnlineEventFields
	}
	if len(fields) > 0 {
		q.Set("fields", strings.Join(fields, ","))
	}

	body, err := client.Get("/api/v1/devices/"+deviceID+"/online-events-list", q)
	if err != nil {
		return err
	}
	return iostreams.FormatOutput(body, f.IO, output, fields)
}

func runOnlineDaily(f *factory.Factory, deviceID string, opts *onlineOptions, cmd *cobra.Command) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	if opts.After != "" {
		q.Set("after", opts.After)
	}
	if opts.Before != "" {
		q.Set("before", opts.Before)
	}

	output, _ := cmd.Flags().GetString("output")

	body, err := client.Get("/api/v1/devices/"+deviceID+"/offline/daily-stats", q)
	if err != nil {
		return err
	}
	return iostreams.FormatOutput(body, f.IO, output, opts.Fields)
}
