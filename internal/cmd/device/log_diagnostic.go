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

func NewCmdLogDiagnostic(f *factory.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "diagnostic <device-id>",
		Short: "Download diagnostic log from device",
		Long: `Download diagnostic log from a device. The device will be asked to collect and upload
the log, which may take up to a few minutes depending on network conditions.`,
		Example: `  # Download diagnostic log to stdout
  incloud device log diagnostic 507f1f77bcf86cd799439011

  # Save to a file
  incloud device log diagnostic 507f1f77bcf86cd799439011 --file diag.log

  # Pipe to grep
  incloud device log diagnostic 507f1f77bcf86cd799439011 | grep -i error`,
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

			u, err := url.Parse(ctx.Host + "/api/v1/devices/" + deviceID + "/logs/download")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			q := u.Query()
			q.Set("type", "diagnostic")
			q.Set("fetchRealtime", "true")
			u.RawQuery = q.Encode()

			fmt.Fprintln(f.IO.ErrOut, "Requesting diagnostic log from device (this may take a few minutes)...")

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

			if file == "" {
				if _, err := io.Copy(f.IO.Out, resp.Body); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}
				return nil
			}

			outFile, err := os.Create(file)
			if err != nil {
				return fmt.Errorf("creating file: %w", err)
			}
			defer func() { _ = outFile.Close() }()

			n, err := io.Copy(outFile, resp.Body)
			if err != nil {
				return fmt.Errorf("writing file: %w", err)
			}

			fmt.Fprintf(f.IO.ErrOut, "Downloaded to %s (%d bytes)\n", file, n)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Output file path (default: stdout)")

	return cmd
}
