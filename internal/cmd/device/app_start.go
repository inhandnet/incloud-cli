package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdAppStart(f *factory.Factory) *cobra.Command {
	var appType string
	var appNames []string

	cmd := &cobra.Command{
		Use:   "start <device-id>",
		Short: "Start applications on a device",
		Long:  "Start container or native applications on an edge device.",
		Example: `  # Start container apps by name
  incloud device app start 507f1f77bcf86cd799439011 --app-type container --app-names myapp,otherapp

  # Start all native apps
  incloud device app start 507f1f77bcf86cd799439011 --app-type native`,
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

			body, err := client.Put("/api/v1/live/devices/"+args[0]+"/apps/start", reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Applications started on device %s.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&appType, "app-type", "", "Application type (e.g. container, native)")
	cmd.Flags().StringSliceVar(&appNames, "app-names", nil, "Application names to start (comma-separated)")
	_ = cmd.MarkFlagRequired("app-type")

	return cmd
}
