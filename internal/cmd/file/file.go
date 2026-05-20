package file

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdFile(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage files",
		Long:  "File operations such as generating pre-signed URLs for uploads.",
	}

	cmd.AddCommand(NewCmdPresign(f))

	return cmd
}
