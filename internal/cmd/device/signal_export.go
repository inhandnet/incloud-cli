package device

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

type signalExportOptions struct {
	After  string
	Before string
	File   string
}

func newCmdSignalExport(f *factory.Factory) *cobra.Command {
	opts := &signalExportOptions{}

	cmd := &cobra.Command{
		Use:   "export <device-id>",
		Short: "Export signal data to Excel",
		Long:  "Export signal quality data as a server-generated Excel file.",
		Example: `  # Export signal data to file
  incloud device signal export 507f1f77bcf86cd799439011 --file signal.xlsx

  # Export with time range
  incloud device signal export 507f1f77bcf86cd799439011 --after 2024-01-01T00:00:00 --before 2024-01-07T00:00:00 --file signal.xlsx

  # Export to stdout (pipe to file)
  incloud device signal export 507f1f77bcf86cd799439011 > signal.xlsx`,
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

			u, err := url.Parse(actx.Host + "/api/v1/devices/" + deviceID + "/signal/export")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			if opts.After != "" {
				q.Set("after", opts.After)
			}
			if opts.Before != "" {
				q.Set("before", opts.Before)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
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

	cmd.Flags().StringVar(&opts.After, "after", "", "Start time (ISO 8601, e.g. 2024-01-01T00:00:00)")
	cmd.Flags().StringVar(&opts.Before, "before", "", "End time (ISO 8601, e.g. 2024-01-02T00:00:00)")
	cmd.Flags().StringVar(&opts.File, "file", "", "Write output to file instead of stdout")

	return cmd
}
