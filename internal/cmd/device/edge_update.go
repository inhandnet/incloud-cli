package device

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdEdgeUpdate(f *factory.Factory) *cobra.Command {
	var envJSON string

	cmd := &cobra.Command{
		Use:   "update <device-id>",
		Short: "Update edge properties of a device",
		Long:  "Update edge-specific properties of a device, such as environment variables.",
		Example: `  # Set environment variables
  incloud device edge update 507f1f77bcf86cd799439011 --env '[{"name":"KEY","value":"val"}]'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			var env []interface{}
			if err := json.Unmarshal([]byte(envJSON), &env); err != nil {
				return fmt.Errorf("invalid --env JSON: %w", err)
			}

			reqBody := map[string]interface{}{
				"env": env,
			}

			output, _ := cmd.Flags().GetString("output")

			body, err := client.Put("/api/v1/live/devices/"+args[0], reqBody)
			if err != nil {
				if body != nil {
					_ = iostreams.FormatOutput(body, f.IO, output)
				}
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Edge device %s updated.\n", args[0])
			return iostreams.FormatOutput(body, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&envJSON, "env", "", `Environment variables as JSON array (e.g. '[{"name":"KEY","value":"val"}]')`)
	_ = cmd.MarkFlagRequired("env")

	return cmd
}
