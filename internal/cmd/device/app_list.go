package device

import (
	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdAppList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device-id>",
		Short:   "List applications on a device",
		Long:    "List container and native applications running on an edge device.",
		Aliases: []string{"ls"},
		Example: `  # List apps on a device
  incloud device app list 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := cmdutil.NewQuery(cmd, nil)
			output, _ := cmd.Flags().GetString("output")

			body, err := client.Get("/api/v1/live/devices/"+args[0]+"/apps", q)
			if err != nil {
				return err
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
