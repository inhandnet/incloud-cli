package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdSignalCurrent(f *factory.Factory) *cobra.Command {
	var fields []string

	cmd := &cobra.Command{
		Use:   "current <device-id>",
		Short: "Show current signal quality",
		Long:  "Display the latest signal quality metrics for a device.",
		Example: `  # Show current signal
  incloud device signal current 507f1f77bcf86cd799439011

  # Table output
  incloud device signal current 507f1f77bcf86cd799439011 -o table`,
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

			u, err := url.Parse(actx.Host + "/api/v1/devices/" + deviceID + "/current-signal")
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), http.NoBody)
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
			case "table":
				flat, err := flattenSignalSeries(body)
				if err != nil {
					return err
				}
				cols := fields
				if len(cols) == 0 && f.IO.IsStdoutTTY() {
					cols = defaultSignalFields
				}
				if err := iostreams.FormatTable(flat, f.IO, cols); err != nil {
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

	cmd.Flags().StringSliceVarP(&fields, "fields", "f", nil, "Fields to display in table mode")

	return cmd
}
