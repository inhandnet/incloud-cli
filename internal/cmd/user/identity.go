package user

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdIdentity(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage user identities across organizations",
		Long:  "View the current user's identities (roles) across different organizations.",
	}

	cmd.AddCommand(NewCmdIdentityList(f))

	return cmd
}
