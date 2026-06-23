package pos

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdMarkedClients(f *factory.Factory) *cobra.Command {
	var level string

	cmd := &cobra.Command{
		Use:   "marked-clients <device-id>",
		Short: "List POS-marked clients on a device",
		Long:  "List clients on a specific device that carry a POS priority level (priority/bypass).",
		Args:  cobra.ExactArgs(1),
		Example: `  # All marked clients on a device
  incloud pos marked-clients 507f1f77bcf86cd799439011

  # Only bypassed clients
  incloud pos marked-clients 507f1f77bcf86cd799439011 --level bypass`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			q := url.Values{}
			if level != "" {
				q.Set("level", level)
			}

			body, err := client.Get("/api/v1/network/devices/"+args[0]+"/marked-clients", q)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&level, "level", "", "Filter by level (priority/bypass)")

	return cmd
}
