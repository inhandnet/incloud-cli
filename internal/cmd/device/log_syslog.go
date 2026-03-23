package device

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type LogSyslogOptions struct {
	After    string
	Before   string
	Keywords []string
	Limit    int
	Fetch    bool
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
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00

  # Actively fetch latest syslog from device (last 15 minutes)
  incloud device log syslog 60af...id --fetch

  # Fetch syslog for a specific time range from device
  incloud device log syslog 60af...id --fetch --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00

  # Filter by keywords
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00 --keywords error --keywords warning`,
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
					before = now.Format("2006-01-02T15:04:05")
				}
				if after == "" {
					// Default to start of today — device uploads its full buffer whose
					// timestamps can span the whole day, not just the last few minutes.
					after = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).
						Format("2006-01-02T15:04:05")
				}
				fmt.Fprintln(f.IO.ErrOut, "Requesting syslog from device (waits up to 40s for device to upload)...")
			}

			q := url.Values{}
			q.Set("startTimestamp", after+"Z")
			q.Set("endTimestamp", before+"Z")
			q.Set("limit", strconv.Itoa(opts.Limit))
			q.Set("index", "0")
			for _, kw := range opts.Keywords {
				q.Add("keywords", kw)
			}

			if opts.Fetch {
				q.Set("fetchRealtime", "true")
			} else {
				q.Set("fetchRealtime", "false")
			}

			body, err := client.Get("/api/v1/devices/"+deviceID+"/logs/download/syslog", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				if json.Valid(body) {
					fmt.Fprintln(f.IO.Out, iostreams.FormatJSON(body, f.IO, output))
				} else {
					fmt.Fprintln(f.IO.Out, string(body))
				}
			default:
				var sr syslogResponse
				if err := json.Unmarshal(body, &sr); err != nil {
					return fmt.Errorf("parsing response: %w", err)
				}
				for _, line := range sr.Result {
					fmt.Fprint(f.IO.Out, line)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time in ISO 8601 format (required without --fetch)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time in ISO 8601 format (required without --fetch)")
	cmd.Flags().StringSliceVar(&opts.Keywords, "keywords", nil, "Filter by keywords")
	cmd.Flags().IntVar(&opts.Limit, "limit", 10000, "Maximum number of log lines")
	cmd.Flags().BoolVar(&opts.Fetch, "fetch", false, "Actively request syslog from device (--after defaults to start of today)")

	return cmd
}
