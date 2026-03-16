package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdPresence(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presence",
		Short: "Device online/offline presence",
		Long:  "View device online/offline presence events and status.",
	}

	cmd.AddCommand(NewCmdPresenceEvents(f))

	return cmd
}
