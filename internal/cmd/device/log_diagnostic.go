package device

import (
	"fmt"
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

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			q.Set("type", "diagnostic")
			q.Set("fetchRealtime", "true")

			fmt.Fprintln(f.IO.ErrOut, "Requesting diagnostic log from device (this may take a few minutes)...")

			body, err := client.Get("/api/v1/devices/"+deviceID+"/logs/download", q)
			if err != nil {
				return err
			}

			if file != "" {
				if err := os.WriteFile(file, body, 0o600); err != nil {
					return fmt.Errorf("writing file: %w", err)
				}
				fmt.Fprintf(f.IO.ErrOut, "Downloaded to %s (%d bytes)\n", file, len(body))
				return nil
			}

			_, err = f.IO.Out.Write(body)
			return err
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Output file path (default: stdout)")

	return cmd
}
