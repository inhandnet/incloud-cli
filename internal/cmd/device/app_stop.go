package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdAppStop(f *factory.Factory) *cobra.Command {
	var appType string
	var appNames []string

	cmd := &cobra.Command{
		Use:   "stop <device-id>",
		Short: "Stop applications on a device",
		Long:  "Stop container or native applications on an edge device.",
		Example: `  # Stop container apps by name
  incloud device app stop 507f1f77bcf86cd799439011 --app-type container --app-names myapp

  # Stop all native apps
  incloud device app stop 507f1f77bcf86cd799439011 --app-type native`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"appType": appType,
			}
			if len(appNames) > 0 {
				reqBody["appNames"] = appNames
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devices/"+args[0]+"/apps/stop", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Applications stopped on device %s.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&appType, "app-type", "", "Application type (e.g. container, native)")
	cmd.Flags().StringSliceVar(&appNames, "app-names", nil, "Application names to stop (comma-separated)")
	_ = cmd.MarkFlagRequired("app-type")

	return cmd
}
