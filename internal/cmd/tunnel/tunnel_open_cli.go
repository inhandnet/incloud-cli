package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTunnelOpenCli(f *factory.Factory) *cobra.Command {
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "open-cli <device-id>",
		Short: "Open a CLI tunnel for a device",
		Long: `Open a remote CLI access tunnel for a device.

Returns a URL for web-based terminal access to the device.
Automatically opens the URL in the default browser (use --no-open to disable).
Each device supports up to 3 concurrent CLI tunnels.
Use 'incloud tunnel close <tunnel-id>' to close the tunnel when done.`,
		Example: `  # Open a CLI tunnel and launch browser
  incloud tunnel open-cli 507f1f77bcf86cd799439011

  # Open without launching browser
  incloud tunnel open-cli 507f1f77bcf86cd799439011 --no-open

  # Capture the URL programmatically
  url=$(incloud tunnel open-cli 507f1f77bcf86cd799439011 --no-open | grep '^URL:' | cut -d' ' -f2)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			endpoint := fmt.Sprintf("/api/v1/ngrok/devices/%s/local-cli", deviceID)
			respBody, err := client.Post(endpoint, nil)
			if err != nil {
				return err
			}

			var resp struct {
				Result struct {
					URL      string `json:"url"`
					Token    string `json:"token"`
					TunnelID string `json:"id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(respBody, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if resp.Result.URL == "" {
				return fmt.Errorf("server returned empty URL: %s", respBody)
			}

			fmt.Fprintf(f.IO.ErrOut, "Tunnel opened (cli) for device %s\n", deviceID)
			fmt.Fprintf(f.IO.Out, "URL: %s\n", resp.Result.URL)
			fmt.Fprintf(f.IO.Out, "Tunnel ID: %s\n", resp.Result.TunnelID)

			if !noOpen {
				if err := browser.OpenURL(resp.Result.URL); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Failed to open browser: %v\n", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the URL in the browser")

	return cmd
}
