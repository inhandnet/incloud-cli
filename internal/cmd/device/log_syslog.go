package device

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

type LogSyslogOptions struct {
	After    string
	Before   string
	Keywords []string
	Limit    int
	Page     int
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
start of today (UTC) and --before defaults to now if not specified.

Each call returns at most --limit lines, oldest first. When more lines match,
a note on stderr gives the returned range, the total, and how to get the rest.
-o json prints {"total": N, "result": [...]}.`,
		Example: `  # Query stored syslog for a time range
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z

  # Actively fetch latest syslog from device (last 15 minutes)
  incloud device log syslog 60af...id --fetch

  # Fetch syslog for a specific time range from device
  incloud device log syslog 60af...id --fetch --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z

  # Fetch with a longer timeout (default 40s)
  incloud device log syslog 60af...id --fetch --timeout 120

  # Get the second batch of 10000 lines
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00Z --before 2024-01-02T00:00:00Z --page 2

  # Filter by keywords
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00Z --before 2024-01-01T01:00:00Z --keywords error --keywords warning`,
		Args: cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "device id", "incloud device list -q %s"),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Page < 1 {
				return fmt.Errorf("--page must be >= 1")
			}
			if opts.Limit < 1 {
				return fmt.Errorf("--limit must be >= 1")
			}
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
			q.Set("startTimestamp", cmdutil.ParseTimeFlag(after))
			q.Set("endTimestamp", cmdutil.ParseTimeFlag(before))
			q.Set("limit", strconv.Itoa(opts.Limit))
			q.Set("index", strconv.Itoa(opts.Page-1))
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

			output, _ := cmd.Flags().GetString("output")
			if output == "json" {
				if sr.Result == nil {
					sr.Result = []string{}
				}
				data, err := json.MarshalIndent(sr, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(f.IO.Out, string(data))
			} else {
				for _, line := range sr.Result {
					fmt.Fprint(f.IO.Out, line)
				}
			}

			printSyslogTruncation(f.IO.ErrOut, sr, opts.Page, opts.Limit)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (e.g. 2025-01-01, 2025-01-01T08:00:00, 2025-01-01T00:00:00Z; required without --fetch)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (e.g. 2025-01-31, 2025-01-31T08:00:00, 2025-01-31T23:59:59Z; required without --fetch)")
	cmd.Flags().StringSliceVar(&opts.Keywords, "keywords", nil, "Filter by keywords")
	cmd.Flags().IntVar(&opts.Limit, "limit", 10000, "Maximum number of log lines per page")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1), each page holds --limit lines")
	cmd.Flags().BoolVar(&opts.Fetch, "fetch", false, "Actively request syslog from device (--after defaults to start of today)")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 40, "Timeout in seconds for device to upload syslog (only with --fetch)")

	return cmd
}

// syslogTimestamp matches the RFC 3164 header the device writes, e.g. "<30> Nov 19 10:48:02".
var syslogTimestamp = regexp.MustCompile(`^<\d+>\s*([A-Z][a-z]{2}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})`)

// printSyslogTruncation tells the caller on stderr when lines beyond this page
// exist, so stdout stays pure log text even when redirected to a file.
func printSyslogTruncation(w io.Writer, sr syslogResponse, page, limit int) {
	first := (page-1)*limit + 1
	last := first + len(sr.Result) - 1
	if len(sr.Result) == 0 || last >= sr.Total {
		return
	}

	where := ""
	if m := syslogTimestamp.FindStringSubmatch(sr.Result[len(sr.Result)-1]); m != nil {
		where = fmt.Sprintf(" (last line at %q)", m[1])
	}
	fmt.Fprintf(w, "\nShowing lines %d-%d of %d%s.\n", first, last, sr.Total, where)
	fmt.Fprintf(w, "Narrow --after/--before, add --keywords, raise --limit, or use --page %d for the next batch.\n", page+1)
}
