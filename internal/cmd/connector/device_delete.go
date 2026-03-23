package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdDeviceDelete(f *factory.Factory) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete <network-id> <device-id> [device-id...]",
		Aliases: []string{"rm"},
		Short:   "Remove devices from a connector network",
		Example: `  # Remove single device
  incloud connector device delete <network-id> <device-id>

  # Remove multiple
  incloud connector device delete <network-id> id1 id2 id3`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			networkID := args[0]
			deviceIDs := args[1:]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			basePath := "/api/v1/connectors/" + networkID + "/devices"
			return deleteConnectorResources(f, client, deviceIDs, yes, "Connector device", basePath, basePath+"/bulk/delete", true)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
