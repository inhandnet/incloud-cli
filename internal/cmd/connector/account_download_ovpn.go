package connector

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type downloadOvpnOptions struct {
	OutPath string
}

func newCmdAccountDownloadOvpn(f *factory.Factory) *cobra.Command {
	opts := &downloadOvpnOptions{}

	cmd := &cobra.Command{
		Use:   "download-ovpn <account-id>",
		Short: "Download OpenVPN configuration file for an account",
		Example: `  # Download to current directory
  incloud connector account download-ovpn <account-id>

  # Specify output path
  incloud connector account download-ovpn <account-id> --out /tmp/client.ovpn`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/connectors/accounts/"+accountID+"/ovpn/download", nil)
			if err != nil {
				return err
			}

			outPath := opts.OutPath
			if outPath == "" {
				outPath = accountID + ".ovpn"
			}

			if dir := filepath.Dir(outPath); dir != "." {
				if err := os.MkdirAll(dir, 0o750); err != nil {
					return err
				}
			}
			if err := os.WriteFile(outPath, body, 0o600); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "OpenVPN config saved to %s\n", outPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.OutPath, "out", "", "Output file path (default: <account-id>.ovpn)")

	return cmd
}
