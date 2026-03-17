package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdLog(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Device logs",
		Long:  "View and download device logs from the InCloud platform.",
	}

	cmd.AddCommand(NewCmdLogSyslog(f))
	cmd.AddCommand(NewCmdLogDiagnostic(f))
	cmd.AddCommand(NewCmdLogMqtt(f))

	return cmd
}
