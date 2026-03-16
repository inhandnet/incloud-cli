package config

import (
	"sort"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdListContexts(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "list-contexts",
		Aliases: []string{"ls"},
		Short:   "List all contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			names := make([]string, 0, len(cfg.Contexts))
			for name := range cfg.Contexts {
				names = append(names, name)
			}
			sort.Strings(names)

			tp := iostreams.NewTablePrinter(f.IO.Out, f.IO.IsStdoutTTY())
			tp.AddRow("CURRENT", "NAME", "HOST", "USER")
			for _, name := range names {
				ctx := cfg.Contexts[name]
				current := ""
				if name == cfg.ActiveContextName() {
					current = "*"
				}
				tp.AddRow(current, name, ctx.Host, ctx.User)
			}
			return tp.Render()
		},
	}
}
