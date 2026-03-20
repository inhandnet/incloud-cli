package config

import (
	"fmt"

	"github.com/spf13/cobra"

	cfgpkg "github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdSetContext(f *factory.Factory) *cobra.Command {
	var host, org string

	cmd := &cobra.Command{
		Use:   "set-context <name>",
		Short: "Create or update a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			ctx, exists := cfg.Contexts[name]
			if !exists {
				ctx = &cfgpkg.Context{}
			}
			ctx.Host = host
			if org != "" {
				ctx.Org = org
			}
			cfg.SetContext(name, ctx)

			if err := f.SaveConfig(); err != nil {
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
	cmd.Flags().StringVar(&org, "org", "", "Organization ID")

	_ = cmd.MarkFlagRequired("host")

	return cmd
}
