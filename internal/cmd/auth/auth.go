package auth

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdAuth(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}

	cmd.AddCommand(NewCmdLogin(f))
	cmd.AddCommand(NewCmdLogout(f))
	cmd.AddCommand(NewCmdStatus(f))

	return cmd
}
