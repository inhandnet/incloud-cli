package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdGroupFirmwares(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firmwares <group-id>",
		Short: "List firmware versions in a device group",
		Args:  cmdutil.ObjectIDArgs(cobra.ExactArgs(1), 0, "device group id", "incloud device group list --name %s"),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get("/api/v1/devicegroups/"+args[0]+"/firmware-versions", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
