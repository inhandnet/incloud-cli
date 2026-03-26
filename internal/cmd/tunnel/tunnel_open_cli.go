package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTunnelOpenCli(f *factory.Factory) *cobra.Command {
	var (
		noOpen    bool
		forward   bool
		localPort int
		ngrokPort int
	)

	cmd := &cobra.Command{
		Use:   "open-cli <device-id>",
		Short: "Open a CLI tunnel for a device",
		Long: `Open a remote CLI access tunnel for a device.

By default, opens the web-based terminal in the browser.
Use --forward to start a local TCP port forward instead.
Use 'incloud tunnel close <tunnel-id>' to close the tunnel when done.`,
		Example: `  # Open a CLI tunnel in the browser
  incloud tunnel open-cli 507f1f77bcf86cd799439011

  # Forward to a local port for ssh/telnet access
  incloud tunnel open-cli 507f1f77bcf86cd799439011 --forward
  incloud tunnel open-cli 507f1f77bcf86cd799439011 --forward --port 2222

  # Just create the tunnel without opening anything
  incloud tunnel open-cli 507f1f77bcf86cd799439011 --no-open`,
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

			fmt.Fprintf(f.IO.ErrOut, "Tunnel opened (cli) for device %s\n", deviceID)
			fmt.Fprintf(f.IO.Out, "Tunnel ID: %s\n", resp.Result.TunnelID)

			if forward {
				opts := &forwardOptions{
					localPort: localPort,
					tunnelID:  resp.Result.TunnelID,
					token:     resp.Result.Token,
					ngrokAddr: fmt.Sprintf("%s:%d", actx.NgrokHost(), ngrokPort),
				}
				// Close tunnel on exit since we created it
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

			if resp.Result.URL != "" {
				fmt.Fprintf(f.IO.Out, "URL: %s\n", resp.Result.URL)
			}
			if resp.Result.Token != "" {
				fmt.Fprintf(f.IO.Out, "Token: %s\n", resp.Result.Token)
			}

			if !noOpen && resp.Result.URL != "" {
				if err := browser.OpenURL(resp.Result.URL); err != nil {
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
