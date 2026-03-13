package cmd

import (
	"github.com/spf13/cobra"
)

type RootOptions struct {
	Output  string
	Context string
}

func NewCmdRoot(version string) *cobra.Command {
	opts := &RootOptions{}

	cmd := &cobra.Command{
		Use:           "incloud",
		Short:         "InCloud Platform CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "", "Output format: json")
	cmd.PersistentFlags().StringVar(&opts.Context, "context", "", "Override active context")

	return cmd
}
