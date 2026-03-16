package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdDeleteContext(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete-context <name>",
		Short: "Delete a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			if _, ok := cfg.Contexts[name]; !ok {
				return fmt.Errorf("context %q not found", name)
			}
			cfg.DeleteContext(name)
			if err := f.SaveConfig(); err != nil {
				return err
			}
			fmt.Fprintf(f.IO.Out, "Deleted context %q\n", name)
			return nil
		},
	}
}
