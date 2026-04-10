package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type LogLocalOptions struct {
	Lines   int
	Path    string
	All     bool
	Timeout int
	File    string
}

func NewCmdLogLocal(f *factory.Factory) *cobra.Command {
	opts := &LogLocalOptions{}

	cmd := &cobra.Command{
		Use:   "local <device-id>",
		Short: "Read log files directly from the device (requires device online)",
		Long: `Read log files directly from the device filesystem in real time.

The device must be online. This command sends a direct method to the device to read
its local log files and prints the content to stdout.`,
		Example: `  # Read the last 100 lines (default)
  incloud device log local 507f1f77bcf86cd799439011

  # Read the last 50 lines
  incloud device log local 507f1f77bcf86cd799439011 --lines 50

  # Read a specific log file
  incloud device log local 507f1f77bcf86cd799439011 --path /var/log/messages

  # Get full log content
  incloud device log local 507f1f77bcf86cd799439011 --all

  # Save to a file for repeated access
  incloud device log local 507f1f77bcf86cd799439011 --file /tmp/device.log

  # With a longer timeout for slow connections
  incloud device log local 507f1f77bcf86cd799439011 --timeout 60`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if opts.All {
				q.Set("all", "true")
			} else if opts.Lines > 0 {
				q.Set("lines", strconv.Itoa(opts.Lines))
			}
			if opts.Path != "" {
				q.Set("localPath", opts.Path)
			}
			q.Set("timeout", strconv.Itoa(opts.Timeout))

			body, err := client.Get("/api/v1/devices/"+deviceID+"/logs/local", q)
			if err != nil {
				return err
			}

			content, err := extractLocalLogs(body)
			if err != nil {
				return err
			}

			if opts.File != "" {
				if err := os.WriteFile(opts.File, content, 0o600); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				absPath, err := filepath.Abs(opts.File)
				if err != nil {
					absPath = opts.File
				}
				fmt.Fprintf(f.IO.ErrOut, "Saved to %s (%d bytes)\n", absPath, len(content))
				return nil
			}

			_, err = f.IO.Out.Write(content)
			return err
		},
	}

	cmd.Flags().IntVar(&opts.Lines, "lines", 100, "Number of log lines to read")
	cmd.Flags().StringVar(&opts.Path, "path", "", "Log file path on the device (e.g. /var/log/messages)")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Get full log content")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", 30, "Timeout in seconds for device response")
	cmd.Flags().StringVar(&opts.File, "file", "", "Save log content to a file instead of stdout")
	cmd.MarkFlagsMutuallyExclusive("all", "lines")

	return cmd
}

// localLogResponse matches the DeviceLog structure from the API.
type localLogResponse struct {
	Status string          `json:"status"`
	Error  string          `json:"error"`
	Result json.RawMessage `json:"result"`
}

type localLogResult struct {
	Logs []string `json:"logs"`
	URL  string   `json:"url"`
}

func extractLocalLogs(body []byte) ([]byte, error) {
	var resp localLogResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if resp.Status != "succeeded" {
		errMsg := resp.Error
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("device returned status %q: %s", resp.Status, errMsg)
	}

	if resp.Result == nil {
		return nil, nil
	}

	var result localLogResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return append(resp.Result, '\n'), nil
	}

	// Case 1: logs array returned (server already downloaded from S3).
	if len(result.Logs) > 0 {
		content := strings.Join(result.Logs, "\n")
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return []byte(content), nil
	}

	// Case 2: presigned URL returned (all=true). Download content.
	if result.URL != "" {
		return downloadContent(result.URL)
	}

	// Case 3: no logs/url — return raw result as fallback.
	return append(resp.Result, '\n'), nil
}

var downloadClient = &http.Client{Timeout: 5 * time.Minute}

func downloadContent(rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("downloading log content: %w", err)
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading log content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading log content: HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
