package config

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdSetContext(f *factory.Factory) *cobra.Command {
	var host string

	cmd := &cobra.Command{
		Use:   "set-context <name>",
		Short: "Create or update a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			var exists bool
			err := f.UpdateConfig(func(cfg *cfgpkg.Config) error {
				ctx, ok := cfg.Contexts[name]
				exists = ok
				if !ok {
					ctx = &cfgpkg.Context{}
				}
				ctx.Host = host
				cfg.SetContext(name, ctx)
				return nil
			})
			if err != nil {
				return err
			}

			action := "Created"
			if exists {
				action = "Updated"
			}
			fmt.Fprintf(f.IO.Out, "%s context %q (%s)\n", action, name, host)
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Platform host URL (required)")
	_ = cmd.MarkFlagRequired("host")

	return cmd
}
