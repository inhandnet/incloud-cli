package firmware

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdFirmware(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firmware",
		Short: "Manage firmwares",
		Long:  "List and inspect firmware versions on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdJob(f))

	return cmd
}
