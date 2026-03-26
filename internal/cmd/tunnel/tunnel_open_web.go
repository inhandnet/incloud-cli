package tunnel

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTunnelOpenWeb(f *factory.Factory) *cobra.Command {
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "open-web <device-id>",
		Short: "Open a Web UI tunnel for a device",
		Long: `Open a remote Web UI access tunnel for a device.

Returns an HTTPS URL that provides browser-based access to the device's web interface.
Automatically opens the URL in the default browser (use --no-open to disable).
Use 'incloud tunnel close <tunnel-id>' to close the tunnel when done.`,
		Example: `  # Open a web tunnel and launch browser
  incloud tunnel open-web 507f1f77bcf86cd799439011

  # Open without launching browser
  incloud tunnel open-web 507f1f77bcf86cd799439011 --no-open`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			endpoint := fmt.Sprintf("/api/v1/ngrok/devices/%s/local-web", deviceID)
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

			// Build a directly usable URL with the auth token
			openURL := resp.Result.URL
			if resp.Result.Token != "" {
				u, err := url.Parse(resp.Result.URL)
				if err == nil {
					q := u.Query()
					q.Set("token", resp.Result.Token)
					u.RawQuery = q.Encode()
					openURL = u.String()
				}
			}

			fmt.Fprintf(f.IO.ErrOut, "Tunnel opened (web) for device %s\n", deviceID)
			fmt.Fprintf(f.IO.Out, "URL: %s\n", openURL)
			fmt.Fprintf(f.IO.Out, "Tunnel ID: %s\n", resp.Result.TunnelID)

			if !noOpen {
				if err := browser.OpenURL(openURL); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Failed to open browser: %v\n", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the URL in the browser")

	return cmd
}
