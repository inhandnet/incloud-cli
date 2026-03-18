package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdShadow(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shadow",
		Short: "Device shadow documents",
		Long:  "Manage device shadow documents (AWS IoT Named Shadows) for desired/reported state.",
	}

	cmd.AddCommand(newCmdShadowList(f))
	cmd.AddCommand(newCmdShadowGet(f))
	cmd.AddCommand(newCmdShadowUpdate(f))
	cmd.AddCommand(newCmdShadowDelete(f))

	return cmd
}
