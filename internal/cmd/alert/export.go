package alert

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type ExportOptions struct {
	After    string
	Before   string
	Status   string
	Priority int
	Device   string
	Group    string
	Type     []string
	Ack      string
	Query    string
	File     string
}

func NewCmdExport(f *factory.Factory) *cobra.Command {
	opts := &ExportOptions{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export alerts",
		Long:  "Export alerts as a server-generated file (CSV). Supports the same filters as 'alert list'.",
		Example: `  # Export all alerts to stdout
  incloud alert export

  # Export to a file
  incloud alert export -f alerts.csv

  # Export unacknowledged alerts
  incloud alert export --ack false -f unacked.csv

  # Export alerts within a time range
  incloud alert export --after 2024-01-01T00:00:00 --before 2024-01-31T23:59:59

  # Pipe to other commands
  incloud alert export | head -20`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			u, err := url.Parse(ctx.Host + "/api/v1/alerts/export")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			var priority *int
			if cmd.Flags().Changed("priority") {
				priority = &opts.Priority
			}
			applyProbeParams(q, opts.After, opts.Before, opts.Status, priority, opts.Device, opts.Group, opts.Type, opts.Ack, opts.Query)
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

			if resp.StatusCode >= 400 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
			}

			w := f.IO.Out
			if opts.File != "" {
				file, err := os.Create(opts.File)
				if err != nil {
					return fmt.Errorf("creating file: %w", err)
				}
				defer func() { _ = file.Close() }()
				w = file
			}

			n, err := io.Copy(w, resp.Body)
			if err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			if opts.File != "" {
				fmt.Fprintf(f.IO.Out, "Exported to %s (%d bytes)\n", opts.File, n)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.After, "after", "", "Filter alerts after this time (e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "Filter alerts before this time (e.g. 2024-01-31T23:59:59)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status (ACTIVE/CLOSED)")
	cmd.Flags().IntVar(&opts.Priority, "priority", 0, "Filter by priority level")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Filter by device ID")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Filter by device group ID")
	cmd.Flags().StringArrayVar(&opts.Type, "type", nil, "Filter by alert type (can be repeated)")
	cmd.Flags().StringVar(&opts.Ack, "ack", "", "Filter by acknowledgement status (true/false)")
	cmd.Flags().StringVarP(&opts.Query, "query", "q", "", "Search by entity name (fuzzy match)")
	cmd.Flags().StringVarP(&opts.File, "file", "f", "", "Write output to file instead of stdout")

	return cmd
}
