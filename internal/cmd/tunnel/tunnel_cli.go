package tunnel

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

const DefaultShellPrompt = `\S+\s*[#$]\s*$`

type CliOptions struct {
	TunnelID      string
	Token         string
	User          string
	Password      string
	Command       string
	ShellMode     bool
	ShellPassword string
	ShellPrompt   string
	Timeout       time.Duration
	NgrokPort     int
}

func NewCmdTunnelCli(f *factory.Factory) *cobra.Command {
	var (
		token         string
		user          string
		password      string
		shellMode     bool
		shellPassword string
		shellPrompt   string
		timeout       int
		ngrokPort     int
	)

	cmd := &cobra.Command{
		Use:   "cli <tunnel-id> <command>",
		Short: "Run a device CLI command through a tunnel",
		Long: `Login to a device through an existing CLI tunnel, execute a command, and return clean output.

This command connects directly to the device's CLI (via telnet over tunnel) and runs
an INOS or shell command. Use this for information not exposed through the platform API,
such as INOS running-config, routing tables, or system-level diagnostics.

For platform-level operations (ping, reboot, packet capture, etc.), prefer 'device exec'
which works through the API without requiring a tunnel.

Requires a tunnel created by 'tunnel open-cli'. The tunnel is NOT closed after execution,
allowing multiple cli calls against the same tunnel for multi-round diagnostics.

Supports INOS CLI commands (default) and BusyBox shell mode (--shell).
In shell mode, use && or ; to chain multiple commands in one call.
Output is cleaned: telnet negotiation, ANSI escapes, command echo,
and prompt are stripped. stdout contains only the command output.`,
		Example: `  # Create a tunnel first
  incloud tunnel open-cli <device-id> --no-open

  # Run an INOS CLI command
  incloud tunnel cli <tunnel-id> --token <token> --user adm --password 123456 "show interface"

  # Run a shell command via BusyBox
  incloud tunnel cli <tunnel-id> --token <token> --user adm --password 123456 \
    --shell --shell-password xxx "uname -a && free -m && ip route"

  # With longer timeout for slow commands (e.g. ping)
  incloud tunnel cli <tunnel-id> --token <token> --user adm --password 123456 \
    --timeout 60 "ping 8.8.8.8"

  # Close the tunnel when done
  incloud tunnel close <tunnel-id>`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			opts := &CliOptions{
				TunnelID:      args[0],
				Token:         token,
				User:          user,
				Password:      password,
				Command:       args[1],
				ShellMode:     shellMode,
				ShellPassword: shellPassword,
				ShellPrompt:   shellPrompt,
				Timeout:       time.Duration(timeout) * time.Second,
				NgrokPort:     ngrokPort,
			}

			ngrokAddr := fmt.Sprintf("%s:%d", actx.NgrokHost(), opts.NgrokPort)
			return RunCli(f, ngrokAddr, opts)
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Auth token for the tunnel (from open-cli output)")
	cmd.Flags().StringVar(&user, "user", "", "Login username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Login password (required)")
	cmd.Flags().BoolVar(&shellMode, "shell", false, "Enter BusyBox shell mode (via 'inhand' command)")
	cmd.Flags().StringVar(&shellPassword, "shell-password", "", "Password for shell mode (required with --shell)")
	cmd.Flags().StringVar(&shellPrompt, "shell-prompt", DefaultShellPrompt, "Shell prompt regex pattern")
	cmd.Flags().IntVar(&timeout, "timeout", 30, "Command timeout in seconds")
	cmd.Flags().IntVar(&ngrokPort, "ngrok-port", defaultNgrokPort, "Ngrok TCP proxy port")

	cmd.MarkFlagRequired("user")
	cmd.MarkFlagRequired("password")

	return cmd
}

// RunCli connects to a tunnel and executes a CLI command on the device.
// Exported so that 'device exec cli' can reuse it.
func RunCli(f *factory.Factory, ngrokAddr string, opts *CliOptions) error {
	// Auto-fetch token from API if not provided
	if opts.Token == "" {
		token, err := fetchTunnelToken(f, opts.TunnelID)
		if err != nil {
			return fmt.Errorf("get tunnel token: %w", err)
		}
		opts.Token = token
	}

	session, err := dialMuxSession(ngrokAddr, opts.TunnelID, opts.Token)
	if err != nil {
		return fmt.Errorf("connect tunnel: %w", err)
	}
	defer session.Close()

	stream, err := session.OpenStream()
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer stream.Close()

	tc := &telnetClient{conn: stream}
	inosPrompts, err := telnetLogin(tc, opts.User, opts.Password, opts.Timeout)
	if err != nil {
		return err
	}

	var execErr error
	if opts.ShellMode {
		execErr = runShellMode(f, tc, opts, inosPrompts)
	} else {
		execErr = runINOSMode(f, tc, opts, inosPrompts)
	}

	telnetExit(tc)

	return execErr
}

// hostnameRE extracts hostname from INOS prompt like "HH:MM:SS hostname#" or "N hostname#"
var hostnameRE = regexp.MustCompile(`(?:\d[\d:]*\s+)?(\S+)[#>]\s*$`)

// telnetLogin performs the telnet login sequence and returns INOS prompt patterns.
func telnetLogin(tc *telnetClient, user, password string, timeout time.Duration) ([]*regexp.Regexp, error) {
	// Wait for telnet negotiation to settle, then send CR to trigger login prompt
	time.Sleep(2 * time.Second)
	if err := tc.write("\r"); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	text, idx := tc.readUntilLiteral([]string{"login:", "Login:"}, timeout)
	if idx == -1 {
		return nil, fmt.Errorf("login failed: no login prompt (got: %s)", truncate(cleanOutput(text), 200))
	}

	if err := tc.write(user + "\r"); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	_, idx = tc.readUntilLiteral([]string{"assword:"}, timeout)
	if idx == -1 {
		return nil, fmt.Errorf("login failed: no password prompt")
	}

	if err := tc.write(password + "\r"); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	text, idx = tc.readUntilLiteral([]string{"# ", "> "}, timeout)
	if idx == -1 {
		cleaned := cleanOutput(text)
		if strings.Contains(cleaned, "Login incorrect") || strings.Contains(cleaned, "Authentication failed") {
			return nil, fmt.Errorf("login failed: incorrect username or password")
		}
		return nil, fmt.Errorf("login failed: no prompt after authentication")
	}

	// Extract hostname from banner for accurate prompt matching
	cleaned := cleanOutput(text)
	prompts := []*regexp.Regexp{
		regexp.MustCompile(`\S+[#>]\s*$`), // fallback
	}
	if m := hostnameRE.FindStringSubmatch(cleaned); m != nil {
		hostname := regexp.QuoteMeta(m[1])
		prompts = []*regexp.Regexp{
			regexp.MustCompile(hostname + `[#>]\s*$`),
		}
	}

	time.Sleep(300 * time.Millisecond)
	return prompts, nil
}

// runINOSMode executes a command in INOS CLI mode.
func runINOSMode(f *factory.Factory, tc *telnetClient, opts *CliOptions, prompts []*regexp.Regexp) error {
	output, err := execINOSCommand(tc, opts.Command, prompts, opts.Timeout)
	if output != "" {
		fmt.Fprint(f.IO.Out, output)
	}
	return err
}

// runShellMode enters BusyBox shell, executes a command, then exits back to INOS.
func runShellMode(f *factory.Factory, tc *telnetClient, opts *CliOptions, inosPrompts []*regexp.Regexp) error {
	if opts.ShellPassword == "" {
		return fmt.Errorf("--shell-password is required with --shell")
	}

	shellPromptRE, err := regexp.Compile(opts.ShellPrompt)
	if err != nil {
		return fmt.Errorf("invalid --shell-prompt regex: %w", err)
	}
	shellPrompts := []*regexp.Regexp{shellPromptRE}

	tc.write("inhand\r")
	_, idx := tc.readUntilLiteral([]string{"assword:", "password:"}, opts.Timeout)
	if idx == -1 {
		return fmt.Errorf("shell mode: no password prompt from 'inhand' command")
	}

	tc.write(opts.ShellPassword + "\r")
	text, idx := tc.readUntil(shellPrompts, opts.Timeout)
	if idx == -1 || strings.Contains(text, "Bad password") {
		return fmt.Errorf("shell mode: incorrect shell password")
	}

	output, err := execShellCommand(tc, opts.Command, shellPrompts, opts.Timeout)
	if output != "" {
		fmt.Fprint(f.IO.Out, output)
	}

	// Exit shell back to INOS
	tc.write("exit\r")
	tc.readUntil(inosPrompts, 10*time.Second)

	return err
}

// telnetExit performs the two-layer INOS exit sequence.
func telnetExit(tc *telnetClient) {
	tc.write("exit\r")
	exitPatterns := []*regexp.Regexp{
		regexp.MustCompile(`Y\|N`),
		regexp.MustCompile(`>\s*$`),
	}
	_, idx := tc.readUntil(exitPatterns, 10*time.Second)
	switch idx {
	case 0:
		// Got Y|N confirmation prompt
		tc.write("Y\r")
	case 1:
		// Dropped from privilege to user mode, exit again
		tc.write("exit\r")
		if _, idx := tc.readUntil(exitPatterns[:1], 10*time.Second); idx == 0 {
			tc.write("Y\r")
		}
	}
	time.Sleep(500 * time.Millisecond)
}

// fetchTunnelToken retrieves the visitor JWT for a tunnel from the API.
func fetchTunnelToken(f *factory.Factory, tunnelID string) (string, error) {
	client, err := f.APIClient()
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("/api/v1/ngrok/tunnels/%s", tunnelID)
	body, err := client.Get(endpoint, nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		Result struct {
			Token string `json:"token"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse tunnel response: %w", err)
	}
	if resp.Result.Token == "" {
		return "", fmt.Errorf("tunnel %s has no token", tunnelID)
	}
	return resp.Result.Token, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
