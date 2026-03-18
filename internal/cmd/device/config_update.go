package device

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdConfigUpdate(f *factory.Factory) *cobra.Command {
	var (
		module  string
		payload string
		file    string
	)

	cmd := &cobra.Command{
		Use:   "update <device-id>",
		Short: "Update device configuration",
		Long:  "Directly update a device's configuration using a JSON merge patch. This creates a session, applies the patch, and commits in one step.",
		Example: `  # Update hostname
  incloud device config update 507f1f77bcf86cd799439011 \
    --payload '{"system":{"hostname":"new-name"}}'

  # Update from a file
  incloud device config update 507f1f77bcf86cd799439011 --file config-patch.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			var data []byte
			var err error

			switch {
			case payload != "" && file != "":
				return fmt.Errorf("--payload and --file are mutually exclusive")
			case payload != "":
				data = []byte(payload)
			case file != "":
				data, err = os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
			default:
				return fmt.Errorf("either --payload or --file is required")
			}

			// Validate JSON
			var body json.RawMessage
			if err := json.Unmarshal(data, &body); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			path := "/api/v1/config/direct?deviceId=" + deviceID
			if module != "" {
				path += "&module=" + module
			}

			resp, err := client.Put(path, body)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" || output == "table" {
				fmt.Fprintf(f.IO.ErrOut, "Configuration updated for device %s.\n", deviceID)
				return nil
			}
			return iostreams.FormatOutput(resp, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&module, "module", "", "Module name (defaults to 'default' on the server)")
	cmd.Flags().StringVar(&payload, "payload", "", "JSON merge patch payload")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file containing the merge patch")

	return cmd
}
