package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func NewCmdCompatibility(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compatibility <device-id>",
		Short: "Check device compatibility support",
		Long:  "Query which capabilities a device supports based on its product and firmware version.",
		Example: `  # Check all compatibilities for a device
  incloud device compatibility 507f1f77bcf86cd799439011

  # Show only supported compatibilities
  incloud device compatibility 507f1f77bcf86cd799439011 --supported

  # JSON output
  incloud device compatibility 507f1f77bcf86cd799439011 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// Step 1: fetch all enabled compatibilities
			q := url.Values{}
			q.Set("enabled", "true")
			q.Set("limit", "100")
			listBody, err := client.Get("/api/v1/product-compatibilities", q)
			if err != nil {
				return fmt.Errorf("failed to list compatibilities: %w", err)
			}

			var listResp struct {
				Result []struct {
					ID string `json:"_id"`
				} `json:"result"`
			}
			if err := json.Unmarshal(listBody, &listResp); err != nil {
				return fmt.Errorf("failed to parse compatibilities: %w", err)
			}

			if len(listResp.Result) == 0 {
				fmt.Fprintln(f.IO.ErrOut, "No compatibilities defined")
				return nil
			}

			compatIDs := make([]string, len(listResp.Result))
			for i, c := range listResp.Result {
				compatIDs[i] = c.ID
			}

			// Step 2: bulk validate for this device
			reqBody := map[string]interface{}{
				"ids":             []string{deviceID},
				"compatibilities": compatIDs,
				"type":            "DEVICE",
			}

			body, err := client.Post("/api/v1/product-compatibilities/bulk-validate", reqBody)
			if err != nil {
				return err
			}

			supported, _ := cmd.Flags().GetBool("supported")
			output, _ := cmd.Flags().GetString("output")
			return iostreams.FormatOutput(body, f.IO, output,
				iostreams.WithTransform(func(data []byte) ([]byte, error) {
					return transformDeviceCompatibility(data, supported)
				}),
			)
		},
	}

	cmd.Flags().Bool("supported", false, "Show only supported compatibilities")

	return cmd
}

// transformDeviceCompatibility flattens the bulk-validate response into a table-friendly format.
// When supportedOnly is true, only entries with "support":true are included.
func transformDeviceCompatibility(data []byte, supportedOnly bool) ([]byte, error) {
	var resp struct {
		Result []struct {
			Compatibilities map[string]json.RawMessage `json:"compatibilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return data, nil
	}

	if len(resp.Result) == 0 {
		return []byte("[]"), nil
	}

	var rows []json.RawMessage
	for _, v := range resp.Result[0].Compatibilities {
		if supportedOnly {
			var item struct {
				Support bool `json:"support"`
			}
			if err := json.Unmarshal(v, &item); err == nil && !item.Support {
				continue
			}
		}
		rows = append(rows, v)
	}

	out, err := json.Marshal(rows)
	if err != nil {
		return data, nil
	}
	return out, nil
}
