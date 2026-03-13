package version

import (
	"fmt"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/spf13/cobra"
)

func NewCmdVersion(f *factory.Factory, version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(f.IO.Out, "incloud version %s\n", version)
		},
	}
}
