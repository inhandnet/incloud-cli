package touch

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdTouch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "touch",
		Short: "Manage DeviceTouch remote access",
		Long:  "Manage remote access clients and connections to downstream equipment via edge devices.",
	}

	cmd.AddCommand(newCmdClient(f))
	cmd.AddCommand(newCmdConnection(f))

	return cmd
}
