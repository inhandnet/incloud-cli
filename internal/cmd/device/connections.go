package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type ConnectionsOptions struct {
	Page   int
	Limit  int
	After  string
	Before string
	Fields []string
}

var defaultConnectionsFields = []string{"timestamp", "eventType", "ipAddress", "disconnectReason"}

func NewCmdConnections(f *factory.Factory) *cobra.Command {
	opts := &ConnectionsOptions{}

	cmd := &cobra.Command{
		Use:   "connections <device-id>",
		Short: "List device connection history",
		Long:  "List online/offline connection events for a specific device.",
		Example: `  # List connection events for a device
  incloud device connections 68d4f9818e517662696751ec

  # With pagination
  incloud device connections 68d4f9818e517662696751ec --page 2 --limit 10

  # Filter by time range
  incloud device connections 68d4f9818e517662696751ec --after 2025-01-01T00:00:00 --before 2025-12-31T23:59:59

  # Table output with selected fields
  incloud device connections 68d4f9818e517662696751ec -o table -f timestamp -f eventType`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID + "/online-events-list")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
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
			if len(fields) == 0 && output == "table" && f.IO.IsStdoutTTY() {
				fields = defaultConnectionsFields
			}
			if len(fields) > 0 {
				q.Set("fields", strings.Join(fields, ","))
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

			switch output {
			case "table":
				if err := iostreams.FormatTable(body, f.IO, fields); err != nil {
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

	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number (starting from 1)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of items per page")
	cmd.Flags().StringVar(&opts.After, "after", "", "Filter events after this time (ISO 8601)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter events before this time (ISO 8601)")
	cmd.Flags().StringSliceVarP(&opts.Fields, "fields", "f", nil, "Fields to return and display")

	return cmd
}
