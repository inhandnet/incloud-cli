package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgeGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get edge properties of a device",
		Long:  "Get edge-specific properties of a device including project status, environment variables, and CLI configuration.",
		Example: `  # Get edge device info
  incloud device edge get 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devices/"+args[0], q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
