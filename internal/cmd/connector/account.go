package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdAccount(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account",
		Aliases: []string{"acc"},
		Short:   "Manage connector accounts (VPN users)",
	}

	cmd.AddCommand(newCmdAccountList(f))
	cmd.AddCommand(newCmdAccountCreate(f))
	cmd.AddCommand(newCmdAccountUpdate(f))
	cmd.AddCommand(newCmdAccountDelete(f))
	cmd.AddCommand(newCmdAccountDownloadOvpn(f))
	cmd.AddCommand(newCmdAccountEvents(f))
	cmd.AddCommand(newCmdAccountLogs(f))
	cmd.AddCommand(newCmdAccountTendency(f))

	return cmd
}
