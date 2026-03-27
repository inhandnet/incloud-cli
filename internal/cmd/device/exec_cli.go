package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmd/tunnel"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExecCli(f *factory.Factory) *cobra.Command {
	var (
		user          string
		password      string
		shellMode     bool
		shellPassword string
		shellPrompt   string
		timeout       int
		ngrokPort     int
	)

	cmd := &cobra.Command{
		Use:   "cli <device-id> <command>",
		Short: "Run a CLI command directly on a device",
		Long: `Create a tunnel, login to the device CLI, execute a command, and return clean output.

This is a convenience wrapper that automatically manages the tunnel lifecycle:
creates a CLI tunnel, runs the command, then closes the tunnel.

For multi-round diagnostics where you need to run commands across multiple
invocations, use 'tunnel open-cli' + 'tunnel cli' + 'tunnel close' instead.

Supports INOS CLI commands (default) and BusyBox shell mode (--shell).
In shell mode, use && or ; to chain multiple commands in one call.
Output is cleaned: telnet negotiation, ANSI escapes, command echo,
and prompt are stripped. stdout contains only the command output.`,
		Example: `  # Run an INOS CLI command
  incloud device exec cli 507f1f77bcf86cd799439011 --user adm --password 123456 "show interface"

  # Run shell commands via BusyBox (chain with &&)
  incloud device exec cli 507f1f77bcf86cd799439011 --user adm --password 123456 \
    --shell --shell-password xxx "uname -a && free -m && ip route"

  # With longer timeout for slow commands
  incloud device exec cli 507f1f77bcf86cd799439011 --user adm --password 123456 \
    --timeout 60 "ping 8.8.8.8"`,
		Args: cobra.ExactArgs(2),
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

			// Create CLI tunnel
			endpoint := fmt.Sprintf("/api/v1/ngrok/devices/%s/local-cli", deviceID)
			respBody, err := client.Post(endpoint, nil)
			if err != nil {
				return fmt.Errorf("create tunnel: %w", err)
			}

			var resp struct {
				Result struct {
					Token    string `json:"token"`
					TunnelID string `json:"id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(respBody, &resp); err != nil {
				return fmt.Errorf("parse tunnel response: %w", err)
			}

			// Close tunnel when done
			defer func() {
				ep := fmt.Sprintf("/api/v1/ngrok/tunnels/%s", resp.Result.TunnelID)
				if _, err := client.Delete(ep); err != nil {
					fmt.Fprintf(f.IO.ErrOut, "Warning: failed to close tunnel: %v\n", err)
				}
			}()

			ngrokAddr := fmt.Sprintf("%s:%d", actx.NgrokHost(), ngrokPort)

			opts := &tunnel.CliOptions{
				TunnelID:      resp.Result.TunnelID,
				Token:         resp.Result.Token,
				User:          user,
				Password:      password,
				Command:       args[1],
				ShellMode:     shellMode,
				ShellPassword: shellPassword,
				ShellPrompt:   shellPrompt,
				Timeout:       time.Duration(timeout) * time.Second,
			}

			return tunnel.RunCli(f, ngrokAddr, opts)
		},
	}

	cmd.Flags().StringVar(&user, "user", "", "Device login username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Device login password (required)")
	cmd.Flags().BoolVar(&shellMode, "shell", false, "Enter BusyBox shell mode (via 'inhand' command)")
	cmd.Flags().StringVar(&shellPassword, "shell-password", "", "Password for shell mode (required with --shell)")
	cmd.Flags().StringVar(&shellPrompt, "shell-prompt", tunnel.DefaultShellPrompt, "Shell prompt regex pattern")
	cmd.Flags().IntVar(&timeout, "timeout", 30, "Command timeout in seconds")
	cmd.Flags().IntVar(&ngrokPort, "ngrok-port", 4443, "Ngrok TCP proxy port")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")

	return cmd
}
