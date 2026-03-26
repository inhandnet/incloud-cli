package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type LogSyslogOptions struct {
	After    string
	Before   string
	Keywords []string
	Limit    int
	Fetch    bool
	Timeout  int
}

type syslogResponse struct {
	Total  int      `json:"total"`
	Result []string `json:"result"`
}

func NewCmdLogSyslog(f *factory.Factory) *cobra.Command {
	opts := &LogSyslogOptions{}

	cmd := &cobra.Command{
		Use:   "syslog <device-id>",
		Short: "View device syslog",
		Long: `View device syslog from the InCloud platform.

By default, queries syslog already uploaded to the platform (requires --after and --before).
With --fetch, actively requests the device to upload its current syslog; --after defaults to
start of today (UTC) and --before defaults to now if not specified.`,
		Example: `  # Query stored syslog for a time range
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z

  # Actively fetch latest syslog from device (last 15 minutes)
  incloud device log syslog 60af...id --fetch

  # Fetch syslog for a specific time range from device
  incloud device log syslog 60af...id --fetch --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z

  # Fetch with a longer timeout (default 40s)
  incloud device log syslog 60af...id --fetch --timeout 120

  # Filter by keywords
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z --keywords error --keywords warning`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Fetch {
				if opts.After == "" {
					return fmt.Errorf("required flag(s) \"after\" not set (or use --fetch to request from device)")
				}
				if opts.Before == "" {
					return fmt.Errorf("required flag(s) \"before\" not set (or use --fetch to request from device)")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			after := opts.After
			before := opts.Before

			if opts.Fetch {
				now := time.Now().UTC()
				if before == "" {
					before = now.Format(time.RFC3339)
				}
				if after == "" {
					// Default to start of today — device uploads its full buffer whose
					// timestamps can span the whole day, not just the last few minutes.
					after = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).
						Format(time.RFC3339)
				}
				fmt.Fprintf(f.IO.ErrOut, "Requesting syslog from device (waits up to %ds for device to upload)...\n", opts.Timeout)
			}

			q := url.Values{}
			q.Set("startTimestamp", after)
			q.Set("endTimestamp", before)
			q.Set("limit", strconv.Itoa(opts.Limit))
			q.Set("index", "0")
			for _, kw := range opts.Keywords {
				q.Add("keywords", kw)
			}

			if opts.Fetch {
				q.Set("fetchRealtime", "true")
				q.Set("timeout", strconv.Itoa(opts.Timeout))
			} else {
				q.Set("fetchRealtime", "false")
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/logs/download/syslog", q)
			if err != nil {
				return err
			}

			var sr syslogResponse
			if err := json.Unmarshal(body, &sr); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			for _, line := range sr.Result {
				fmt.Fprint(f.IO.Out, line)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time in ISO 8601 format (required without --fetch)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time in ISO 8601 format (required without --fetch)")
	cmd.Flags().StringSliceVar(&opts.Keywords, "keywords", nil, "Filter by keywords")
	cmd.Flags().IntVar(&opts.Limit, "limit", 10000, "Maximum number of log lines")
	cmd.Flags().BoolVar(&opts.Fetch, "fetch", false, "Actively request syslog from device (--after defaults to start of today)")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 40, "Timeout in seconds for device to upload syslog (only with --fetch)")

	return cmd
}