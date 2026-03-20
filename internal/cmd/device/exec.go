package device

import (
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
  speedtest-config   Get speed test configuration
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
	cmd.AddCommand(NewCmdExecSpeedtestConfig(f))
	cmd.AddCommand(NewCmdExecSpeedtestHistory(f))
	cmd.AddCommand(NewCmdExecCapture(f))
	cmd.AddCommand(NewCmdExecCaptureStatus(f))
	cmd.AddCommand(NewCmdExecFlowscan(f))
	cmd.AddCommand(NewCmdExecFlowscanStatus(f))
	cmd.AddCommand(NewCmdExecCancel(f))
	cmd.AddCommand(NewCmdExecInterfaces(f))

	return cmd
}
