package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/build"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdVersion(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(f.IO.Out, "incloud version %s\n", build.Version)
			fmt.Fprintf(f.IO.Out, "commit: %s\n", build.Commit)
			fmt.Fprintf(f.IO.Out, "built:  %s\n", build.Date)
			fmt.Fprintf(f.IO.Out, "go:     %s\n", build.GoVersion())
		},
	}
}
