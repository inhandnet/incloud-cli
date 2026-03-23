package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdEndpointDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <network-id> <endpoint-id> [endpoint-id...]",
		Aliases: []string{"rm"},
		Short:   "Delete endpoints from a connector network",
		Example: `  # Delete single endpoint
  incloud connector endpoint delete 66827b3ccfb1842140f4222f ep123

  # Delete multiple endpoints
  incloud connector endpoint delete 66827b3ccfb1842140f4222f ep1 ep2 ep3

  # Skip confirmation
  incloud connector endpoint delete 66827b3ccfb1842140f4222f ep123 -y`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			networkID := args[0]
			endpointIDs := args[1:]
			basePath := "/api/v1/connectors/" + networkID + "/endpoints"
			bulkPath := basePath + "/bulk/delete"

			return deleteConnectorResources(f, client, endpointIDs, yes, "Connector endpoint", basePath, bulkPath, true)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
