package device

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

type LogSyslogOptions struct {
	After    string
	Before   string
	Keywords []string
	Limit    int
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
		Long:  "Download syslog history for a device within a specified time range.",
		Example: `  # Get syslog for the last hour
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00

  # Filter by keywords
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00 --keywords error --keywords warning

  # Limit results
  incloud device log syslog 60af...id --after 2024-01-01T00:00:00 --before 2024-01-01T01:00:00 --limit 100`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

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

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID + "/logs/download/syslog/history")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			q.Set("startTimestamp", opts.After+"Z")
			q.Set("endTimestamp", opts.Before+"Z")
			q.Set("limit", strconv.Itoa(opts.Limit))
			q.Set("index", "0")
			for _, kw := range opts.Keywords {
				q.Add("keywords", kw)
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
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

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time in ISO 8601 format (required)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time in ISO 8601 format (required)")
	cmd.Flags().StringSliceVar(&opts.Keywords, "keywords", nil, "Filter by keywords")
	cmd.Flags().IntVar(&opts.Limit, "limit", 10000, "Maximum number of log lines")
	_ = cmd.MarkFlagRequired("after")
	_ = cmd.MarkFlagRequired("before")

	return cmd
}
