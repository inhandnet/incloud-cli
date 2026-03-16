package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdCurrentContext(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current-context",
		Short: "Show the current context name",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			name := cfg.ActiveContextName()
			if name == "" {
				return fmt.Errorf("no current context set")
			}
			fmt.Fprintln(f.IO.Out, name)
			return nil
		},
	}
}
