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
	var (
		noOpen    bool
		forward   bool
		localPort int
		ngrokPort int
	)

	cmd := &cobra.Command{
		Use:   "open-web <device-id>",
		Short: "Open a Web UI tunnel for a device",
		Long: `Open a remote Web UI access tunnel for a device.

By default, opens the device's web interface in the browser.
Use --forward to start a local TCP port forward instead.
Use 'incloud tunnel close <tunnel-id>' to close the tunnel when done.`,
		Example: `  # Open a web tunnel in the browser
  incloud tunnel open-web 507f1f77bcf86cd799439011

  # Forward to a local port
  incloud tunnel open-web 507f1f77bcf86cd799439011 --forward --port 8080

  # Just create the tunnel without opening anything
  incloud tunnel open-web 507f1f77bcf86cd799439011 --no-open`,
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

			fmt.Fprintf(f.IO.ErrOut, "Tunnel opened (web) for device %s\n", deviceID)
			fmt.Fprintf(f.IO.Out, "Tunnel ID: %s\n", resp.Result.TunnelID)

			if forward {
				opts := &forwardOptions{
					localPort: localPort,
					tunnelID:  resp.Result.TunnelID,
					token:     resp.Result.Token,
					ngrokAddr: fmt.Sprintf("%s:%d", actx.NgrokHost(), ngrokPort),
				}
				defer func() {
					ep := fmt.Sprintf("/api/v1/ngrok/tunnels/%s", resp.Result.TunnelID)
					if _, err := client.Delete(ep); err != nil {
						fmt.Fprintf(f.IO.ErrOut, "Warning: failed to close tunnel: %v\n", err)
					} else {
						fmt.Fprintf(f.IO.ErrOut, "Tunnel %s closed\n", resp.Result.TunnelID)
					}
				}()
				return runForward(f, opts)
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

			if openURL != "" {
				fmt.Fprintf(f.IO.Out, "URL: %s\n", openURL)
			}
			if resp.Result.Token != "" {
				fmt.Fprintf(f.IO.Out, "Token: %s\n", resp.Result.Token)
			}

			if !noOpen && openURL != "" {
				if err := browser.OpenURL(openURL); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Failed to open browser: %v\n", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not open the URL in the browser")
	cmd.Flags().BoolVar(&forward, "forward", false, "Forward tunnel to a local port instead of opening browser")
	cmd.Flags().IntVarP(&localPort, "port", "p", 0, "Local port for --forward (0 = random)")
	cmd.Flags().IntVar(&ngrokPort, "ngrok-port", 4443, "Ngrok TCP proxy port")
	cmd.MarkFlagsMutuallyExclusive("no-open", "forward")

	return cmd
}
