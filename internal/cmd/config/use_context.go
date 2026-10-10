package config

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdUseContext(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use-context <name>",
		Short: "Switch to a different context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			err := f.UpdateConfig(func(cfg *cfgpkg.Config) error {
				if _, ok := cfg.Contexts[name]; !ok {
					return fmt.Errorf("context %q not found", name)
				}
				cfg.CurrentContext = name
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IO.Out, "Switched to context %q\n", name)
			return nil
		},
	}
}
