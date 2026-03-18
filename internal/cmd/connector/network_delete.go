package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdNetworkDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <id> [id...]",
		Aliases: []string{"rm"},
		Short:   "Delete connector networks",
		Example: `  # Delete single network
  incloud connector network delete 66827b3ccfb1842140f4222f

  # Delete multiple
  incloud connector network delete id1 id2 id3

  # Skip confirmation
  incloud connector network delete 66827b3ccfb1842140f4222f -y`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			return deleteConnectorResources(f, client, args, yes, "Connector network", "/api/v1/connectors", "/api/v1/connectors/bulk/delete")
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
