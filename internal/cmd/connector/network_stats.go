package connector

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newCmdNetworkStats(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show connector statistics overview",
		Example: `  # Show total counts of networks, accounts, devices, endpoints
  incloud connector network stats`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/connectors/statistics", nil)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, body)
		},
	}

	return cmd
}
