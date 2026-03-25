package device

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdLogDiagnostic(f *factory.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "diagnostic <device-id>",
		Short: "Download and decrypt diagnostic log from device",
		Long: `Download diagnostic log from a device, automatically decrypt it, and save as a .tar.gz file.
The device will be asked to collect and upload the log, which may take up to a few minutes
depending on network conditions.`,
		Example: `  # Download to an auto-generated temp file
  incloud device log diagnostic 507f1f77bcf86cd799439011

  # Save to a specific file
  incloud device log diagnostic 507f1f77bcf86cd799439011 --file diag.tar.gz`,
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

			// Decrypt if the response is AES-encrypted.
			if isDiagnosticEncrypted(body) {
				fmt.Fprintln(f.IO.ErrOut, "Decrypting diagnostic log...")
				body, err = decryptDiagnosticLog(body)
				if err != nil {
					return fmt.Errorf("decrypting diagnostic log: %w", err)
				}
			}

			// Determine output path.
			outPath := file
			if outPath == "" {
				tmpFile, err := os.CreateTemp("", "diag-*.tar.gz")
				if err != nil {
					return fmt.Errorf("creating temp file: %w", err)
				}
				outPath = tmpFile.Name()
				_ = tmpFile.Close()
			}

			if err := os.WriteFile(outPath, body, 0o600); err != nil { //nolint:gosec // outPath is either user-provided --file or os.CreateTemp
				return fmt.Errorf("writing file: %w", err)
			}

			absPath, _ := filepath.Abs(outPath)
			fmt.Fprintf(f.IO.ErrOut, "Saved to %s (%d bytes)\n", absPath, len(body))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Output file path (default: auto-generated temp file)")

	return cmd
}
