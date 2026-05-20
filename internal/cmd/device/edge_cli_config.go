package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgeCliConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli-config <device-id>",
		Short: "Get the current CLI configuration from a device",
		Long:  "Retrieve the current running CLI configuration directly from an edge device.",
		Example: `  # Get current CLI config
  incloud device edge cli-config 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devices/"+args[0]+"/cli-config", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
