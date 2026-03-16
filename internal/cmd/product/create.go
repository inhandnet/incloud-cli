package product

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			cfg, err := f.Config()
			if err != nil {
				return err
			}
			ctx, err := cfg.ActiveContext()
			if err != nil {
				return err
			}

			client, err := f.HttpClient()
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

			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			reqURL := ctx.Host + "/api/v1/products"
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, reqURL, bytes.NewReader(jsonBytes))
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

// formatOutput handles -o flag output formatting.
func formatOutput(cmd *cobra.Command, streams *iostreams.IOStreams, body []byte, columns []string) error {
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "table":
		return iostreams.FormatTable(body, streams, columns)
	case "yaml":
		s, err := iostreams.FormatYAML(body)
		if err != nil {
			return err
		}
		fmt.Fprintln(streams.Out, s)
	default:
		if json.Valid(body) {
			fmt.Fprintln(streams.Out, iostreams.FormatJSON(body, streams, output))
		} else {
			fmt.Fprintln(streams.Out, string(body))
		}
	}
	return nil
}
