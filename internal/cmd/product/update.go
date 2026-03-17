package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			actx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
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

			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			reqURL := actx.Host + "/api/v1/products/" + opts.ID
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, reqURL, bytes.NewReader(jsonBytes))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if err := iostreams.FormatOutput(respBody, f.IO, output, nil); err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Description, "description", "", "Product description")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Product status: INDEVELOPMENT or PUBLISHED")
	cmd.Flags().StringVar(&opts.ValidatedField, "validated-field", "", "Validated field: MAC or IMEI")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable, max 100)")

	return cmd
}
