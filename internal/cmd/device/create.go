package device

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
	Name        string
	SN          string
	Product     string
	Description string
	Group       string
	Mac         string
	IMEI        string
	Labels      []string
	Metadata    []string
}

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a new device on the InCloud platform.",
		Example: `  # Create a device with required fields
  incloud device create --name "My Router" --sn "SN12345"

  # With product and description
  incloud device create --name "My Router" --sn "SN12345" --product IR615 --description "Office router"

  # With labels and metadata
  incloud device create --name "My Router" --sn "SN12345" --label env=prod --label region=us
  incloud device create --name "My Router" --sn "SN12345" --metadata key1=val1`,
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
				"name":         opts.Name,
				"serialNumber": opts.SN,
			}
			if opts.Product != "" {
				body["product"] = opts.Product
			}
			if opts.Description != "" {
				body["description"] = opts.Description
			}
			if opts.Group != "" {
				body["devicegroupId"] = opts.Group
			}
			if opts.Mac != "" {
				body["mac"] = opts.Mac
			}
			if opts.IMEI != "" {
				body["imei"] = opts.IMEI
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

			reqURL := ctx.Host + "/api/v1/devices"
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

	cmd.Flags().StringVar(&opts.Name, "name", "", "Device name (required)")
	cmd.Flags().StringVar(&opts.SN, "sn", "", "Serial number (required)")
	cmd.Flags().StringVar(&opts.Product, "product", "", "Product model")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Device description (max 256 chars)")
	cmd.Flags().StringVar(&opts.Group, "group", "", "Device group ID")
	cmd.Flags().StringVar(&opts.Mac, "mac", "", "MAC address (XX:XX:XX:XX:XX:XX)")
	cmd.Flags().StringVar(&opts.IMEI, "imei", "", "IMEI (15-17 digits)")
	cmd.Flags().StringArrayVar(&opts.Labels, "label", nil, "Label in key=value format (repeatable, max 10)")
	cmd.Flags().StringArrayVar(&opts.Metadata, "metadata", nil, "Metadata in key=value format (repeatable)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("sn")

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

// formatOutput handles -o flag output formatting, shared by create/update.
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
