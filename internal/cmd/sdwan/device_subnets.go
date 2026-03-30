package sdwan

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdDeviceSubnets(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "device-subnets <deviceId>",
		Short: "Get subnets for a device",
		Example: `  # View subnets reported by a device
  incloud sdwan device-subnets <deviceId>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Get(apiBase+"/devices/"+args[0]+"/subnets", nil)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
