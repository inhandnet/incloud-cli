package auth

import (
	"fmt"
	"time"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

func NewCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			name := cfg.ActiveContextName()
			if name == "" {
				fmt.Fprintln(f.IO.Out, "No active context")
				return nil
			}

			ctx, ok := cfg.Contexts[name]
			if !ok {
				return fmt.Errorf("context %q not found", name)
			}

			out := f.IO.Out
			fmt.Fprintf(out, "Context:  %s\n", iostreams.Bold(name))
			fmt.Fprintf(out, "Host:     %s\n", ctx.Host)

			if ctx.User != "" {
				fmt.Fprintf(out, "User:     %s\n", ctx.User)
			}
			if ctx.Org != "" {
				fmt.Fprintf(out, "Org:      %s\n", ctx.Org)
			}

			if ctx.Token == "" {
				fmt.Fprintf(out, "Status:   %s\n", iostreams.Red("not logged in"))
			} else if !ctx.ExpiresAt.IsZero() && ctx.ExpiresAt.Before(time.Now()) {
				fmt.Fprintf(out, "Status:   %s\n", iostreams.Yellow("token expired"))
			} else {
				fmt.Fprintf(out, "Status:   %s\n", iostreams.Green("logged in"))
			}

			return nil
		},
	}
}
