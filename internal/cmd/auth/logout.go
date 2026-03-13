package auth

import (
	"fmt"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

func NewCmdLogout(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "logout [context]",
		Short: "Clear authentication tokens",
		Long:  "Remove tokens from a context. Defaults to current context.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name := cfg.ActiveContextName()
			if len(args) > 0 {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("no context specified and no current context set")
			}

			ctx, ok := cfg.Contexts[name]
			if !ok {
				return fmt.Errorf("context %q not found", name)
			}

			ctx.Token = ""
			ctx.RefreshToken = ""

			if err := f.SaveConfig(); err != nil {
				return err
			}

			fmt.Fprintf(f.IO.Out, "%s Logged out of context %q\n", iostreams.Green("✓"), name)
			return nil
		},
	}
}
