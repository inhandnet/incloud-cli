package license

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdType(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "type",
		Short: "Manage license types",
		Long:  "List and view license type definitions on the InCloud platform.",
	}

	cmd.AddCommand(NewCmdTypeList(f))
	cmd.AddCommand(NewCmdTypeGet(f))

	return cmd
}
