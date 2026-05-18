package device

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/cmdutil"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgeCliConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli-config <device-id>",
		Short: "Get the current CLI configuration from a device",
		Long: `Retrieve the current running CLI configuration from an edge device.

If the device is online, the configuration is fetched directly from the device.
If the device is offline, the last cached configuration is returned instead.`,
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
			deviceID := args[0]

			// Try fetching live CLI config directly from the device.
			body, err := client.Get("/api/v1/live/devices/"+deviceID+"/cli-config", q)
			if err != nil {
				// If the live call fails, fall back to the cached config.
				var httpErr *api.HTTPError
				if !errors.As(err, &httpErr) {
					return err
				}

				body, err = client.Get("/api/v1/live/devices/"+deviceID, q)
				if err != nil {
					return err
				}

				var resp struct {
					Result struct {
						CliConfig json.RawMessage `json:"cliConfig"`
					} `json:"result"`
				}
				if err := json.Unmarshal(body, &resp); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}

				if resp.Result.CliConfig == nil {
					return fmt.Errorf("no CLI configuration available for this device")
				}

				body = resp.Result.CliConfig
			}

			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	return cmd
}
