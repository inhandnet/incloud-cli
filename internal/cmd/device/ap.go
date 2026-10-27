package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdAP(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ap",
		Short: "AC-AP networking (APs managed by an AC device)",
		Long:  "Inspect the APs managed by an AC device and their uplink topology.",
	}

	cmd.AddCommand(newCmdAPList(f))

	return cmd
}
