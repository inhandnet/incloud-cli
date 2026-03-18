package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdAccountDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <network-id> <account-id> [account-id...]",
		Aliases: []string{"rm"},
		Short:   "Delete accounts from a connector network",
		Example: `  # Delete single account
  incloud connector account delete <network-id> <account-id>

  # Delete multiple accounts
  incloud connector account delete <network-id> id1 id2 id3

  # Skip confirmation
  incloud connector account delete <network-id> <account-id> -y`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]
			accountIDs := args[1:]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			basePath := "/api/v1/connectors/" + networkID + "/accounts"
			return deleteConnectorResources(f, client, accountIDs, yes, "Connector account", basePath, basePath+"/bulk/delete")
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
