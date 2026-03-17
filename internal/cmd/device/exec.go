package device

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdExec(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute remote methods and diagnostics on devices",
		Long: `Execute remote methods, built-in operations, and diagnostic tools on devices.

Remote methods:
  method             Invoke a custom remote method
  reboot             Reboot a device
  restore-defaults   Restore factory defaults

Diagnostics:
  ping               Ping a host from the device
  traceroute         Traceroute to a host
  speedtest          Run speed test
  speedtest-history  View speed test history
  capture            Start packet capture (tcpdump)
  capture-status     Get capture status
  flowscan           Start flow scan
  flowscan-status    Get flow scan status
  cancel             Cancel a diagnostic task
  interfaces         List network interfaces`,
	}

	// Remote methods
	cmd.AddCommand(NewCmdExecMethod(f))
	cmd.AddCommand(NewCmdExecReboot(f))
	cmd.AddCommand(NewCmdExecRestore(f))

	// Diagnostics
	cmd.AddCommand(NewCmdExecPing(f))
	cmd.AddCommand(NewCmdExecTraceroute(f))
	cmd.AddCommand(NewCmdExecSpeedtest(f))
	cmd.AddCommand(NewCmdExecSpeedtestHistory(f))
	cmd.AddCommand(NewCmdExecCapture(f))
	cmd.AddCommand(NewCmdExecCaptureStatus(f))
	cmd.AddCommand(NewCmdExecFlowscan(f))
	cmd.AddCommand(NewCmdExecFlowscanStatus(f))
	cmd.AddCommand(NewCmdExecCancel(f))
	cmd.AddCommand(NewCmdExecInterfaces(f))

	return cmd
}

// confirmPrompt asks the user for confirmation. Returns true if confirmed.
func confirmPrompt(f *factory.Factory, message string) (bool, error) {
	if file, ok := f.IO.In.(*os.File); !ok || !isatty.IsTerminal(file.Fd()) {
		return false, fmt.Errorf("terminal is non-interactive; use --yes to confirm")
	}
	fmt.Fprintf(f.IO.ErrOut, "%s (y/N) ", message)
	reader := bufio.NewReader(f.IO.In)
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	a := strings.TrimSpace(answer)
	return a == "y" || a == "Y", nil
}
