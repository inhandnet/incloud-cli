package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupProjectDevicesSummary(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "devices-summary <group-id>",
		Short: "Get deployed devices summary for a device group",
		Long:  "Get a summary of deployed project versions across devices in a device group.",
		Example: `  # Get devices summary
  incloud device group project devices-summary 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devicegroups/"+args[0]+"/projects/deployed-devices-summary", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
