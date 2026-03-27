package product

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type UpdateOptions struct {
	ID             string
	Description    string
	Status         string
	ValidatedField string
	Labels         []string
	Metadata       []string
}

func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	opts := &UpdateOptions{}

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a product",
		Long:  "Update an existing product on the InCloud platform.",
		Example: `  # Update description
  incloud product update 507f1f77bcf86cd799439011 --description "Updated"

  # Publish a product
  incloud product update 507f1f77bcf86cd799439011 --status PUBLISHED

  # Update validated field
  incloud product update 507f1f77bcf86cd799439011 --validated-field IMEI

  # Update labels and metadata
  incloud product update 507f1f77bcf86cd799439011 --label env=staging --metadata key1=val1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ID = args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := make(map[string]interface{})

			if cmd.Flags().Changed("description") {
				body["description"] = opts.Description
			}
			if cmd.Flags().Changed("status") {
				body["status"] = opts.Status
			}
			if cmd.Flags().Changed("validated-field") {
				body["validatedField"] = opts.ValidatedField
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
				return fmt.Errorf("no fields to update; specify at least one of --description, --status, --validated-field, --label, or --metadata")
			}

			respBody, err := client.Put("/api/v1/products/"+opts.ID, body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output)
				}
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output)
		},
	}

	cmd.Flags().StringVar(&opts.Description, "description", "", "Product description")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Product status: INDEVELOPMENT or PUBLISHED")
	cmd.Flags().StringVar(&opts.ValidatedField, "validated-field", "", "Validated field: MAC or IMEI")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable, max 100)")

	return cmd
}
