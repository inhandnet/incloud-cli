package device

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			reqURL := actx.Host + "/api/v1/devices/" + opts.ID
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

			if err := formatOutput(cmd, f.IO, respBody, nil); err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				return fmt.Errorf("HTTP %d", resp.StatusCode)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Device name")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Device description (max 256 chars)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable)")

	return cmd
}
