package device

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newCmdShadowUpdate(f *factory.Factory) *cobra.Command {
	var (
		name    string
		payload string
		file    string
	)

	cmd := &cobra.Command{
		Use:   "update <device-id>",
		Short: "Update a shadow document",
		Long:  "Update the desired state of a device shadow document.",
		Example: `  # Update shadow with inline JSON
  incloud device shadow update 507f1f77bcf86cd799439011 --name default \
    --payload '{"state":{"desired":{"temperature":25}}}'

  # Update shadow from a file
  incloud device shadow update 507f1f77bcf86cd799439011 --name default \
    --file shadow.json`,
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

			path := fmt.Sprintf("/api/v1/devices/%s/shadow?name=%s", deviceID, name)
			resp, err := client.Post(path, body)
			if err != nil {
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(resp, f.IO, output, nil,
				iostreams.WithTransform(extractResultArray),
			)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Shadow name (required)")
	cmd.Flags().StringVar(&payload, "payload", "", "Shadow JSON payload")
	cmd.Flags().StringVar(&file, "file", "", "Path to JSON file containing shadow payload")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}
