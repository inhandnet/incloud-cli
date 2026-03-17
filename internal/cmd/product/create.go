package product

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

type CreateOptions struct {
	Name           string
	Type           string
	Description    string
	ValidatedField string
	Labels         []string
	Metadata       []string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a product",
		Long:  "Create a new product on the InCloud platform.",
		Example: `  # Create a product with required fields
  incloud product create --name IR615 --type router

  # With validated field
  incloud product create --name IR615 --type router --validated-field IMEI

  # With description and labels
  incloud product create --name IR615 --type router --description "Edge router" --label env=prod`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"name":           opts.Name,
				"productType":    opts.Type,
				"validatedField": opts.ValidatedField,
			}
			if opts.Description != "" {
				body["description"] = opts.Description
			}
			if len(opts.Labels) > 0 {
				labels, err := parseLabels(opts.Labels)
				if err != nil {
					return err
				}
				body["labels"] = labels
			}
			if len(opts.Metadata) > 0 {
				meta, err := parseKeyValues(opts.Metadata)
				if err != nil {
					return err
				}
				body["metadata"] = meta
			}

			respBody, err := client.Post("/api/v1/products", body)
			if err != nil {
				output, _ := cmd.Flags().GetString("output")
				if respBody != nil {
					_ = iostreams.FormatOutput(respBody, f.IO, output, nil)
				}
				return err
			}

			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(respBody, f.IO, output, nil)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Product name (required, 1-128 chars)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Product type (required)")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Product description")
	cmd.Flags().StringVar(&opts.ValidatedField, "validated-field", "MAC", "Validated field: MAC or IMEI")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable, max 100)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

// parseLabels converts ["key=value", ...] into [{"name":"key","value":"value"}, ...]
func parseLabels(pairs []string) ([]map[string]string, error) {
	labels := make([]map[string]string, 0, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label: %s (expected key=value)", pair)
		}
		labels = append(labels, map[string]string{
			"name":  parts[0],
			"value": parts[1],
		})
	}
	return labels, nil
}

// parseKeyValues converts ["key=value", ...] into {"key":"value", ...}
func parseKeyValues(pairs []string) (map[string]string, error) {
	m := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key=value: %s", pair)
		}
		m[parts[0]] = parts[1]
	}
	return m, nil
}
