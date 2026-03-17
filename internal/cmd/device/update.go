package device

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

type UpdateOptions struct {
	ID          string
	Name        string
	Description string
	Labels      []string
	Metadata    []string
}

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a device",
		Long:  "Update an existing device on the InCloud platform.",
		Example: `  # Update device name
  incloud device update 507f1f77bcf86cd799439011 --name "New Name"

  # Update description
  incloud device update 507f1f77bcf86cd799439011 --description "Updated description"

  # Update labels
  incloud device update 507f1f77bcf86cd799439011 --label env=staging --label region=eu

  # Update metadata
  incloud device update 507f1f77bcf86cd799439011 --metadata key1=val1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ID = args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})

			if cmd.Flags().Changed("name") {
				body["name"] = opts.Name
			}
			if cmd.Flags().Changed("description") {
				body["description"] = opts.Description
			}
			if cmd.Flags().Changed("label") {
				labels, err := parseLabels(opts.Labels)
				if err != nil {
					return err
				}
				body["labels"] = labels
			}
			if cmd.Flags().Changed("metadata") {
				meta, err := parseKeyValues(opts.Metadata)
				if err != nil {
					return err
				}
				body["metadata"] = meta
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update; specify at least one of --name, --description, --label, or --metadata")
			}

			respBody, err := client.Put("/api/v1/devices/"+opts.ID, body)
			if err != nil {
				return err
			}

			return formatOutput(cmd, f.IO, respBody, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Device name")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Device description (max 256 chars)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable)")

	return cmd
}
