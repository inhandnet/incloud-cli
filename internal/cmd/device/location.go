package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdLocation(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location",
		Short: "Manage device location",
		Long:  "View, set, unpin, or refresh device location information.",
	}

	cmd.AddCommand(newCmdLocationGet(f))
	cmd.AddCommand(newCmdLocationSet(f))
	cmd.AddCommand(newCmdLocationUnpin(f))
	cmd.AddCommand(newCmdLocationRefresh(f))

	return cmd
}

func newCmdLocationGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <device-id>",
		Short: "Get device location",
		Long:  "Display the current location information for a device.",
		Example: `  # Get device location
  incloud device location get 507f1f77bcf86cd799439011

  # Table output
  incloud device location get 507f1f77bcf86cd799439011 -o table`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			query := url.Values{}
			query.Set("fields", "location")

			body, err := client.Get("/api/v1/devices/"+deviceID, query)
			if err != nil {
				return err
			}

			locBody, err := extractLocation(body)
			if err != nil {
				return err
			}
			return formatOutput(cmd, f.IO, locBody)
		},
	}

	return cmd
}

type locationSetOptions struct {
	Longitude float64
	Latitude  float64
	Address   string
}

func newCmdLocationSet(f *factory.Factory) *cobra.Command {
	opts := &locationSetOptions{}

	cmd := &cobra.Command{
		Use:   "set <device-id>",
		Short: "Set device location (pin)",
		Long:  "Set a fixed location for a device. This pins the location and disables automatic positioning.",
		Example: `  # Set device location
  incloud device location set 507f1f77bcf86cd799439011 \
    --longitude 119.26 --latitude 30.92 --address "Hangzhou, China"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			payload := map[string]interface{}{
				"location": map[string]interface{}{
					"longitude": opts.Longitude,
					"latitude":  opts.Latitude,
				},
				"address": opts.Address,
				"pinned":  true,
			}

			body, err := client.Put("/api/v1/devices/"+deviceID+"/location", payload)
			if err != nil {
				return err
			}

			locBody, err := extractLocation(body)
			if err != nil {
				return err
			}
			return formatOutput(cmd, f.IO, locBody)
		},
	}

	cmd.Flags().Float64Var(&opts.Longitude, "longitude", 0, "Longitude coordinate (required)")
	cmd.Flags().Float64Var(&opts.Latitude, "latitude", 0, "Latitude coordinate (required)")
	cmd.Flags().StringVar(&opts.Address, "address", "", "Address description (required)")
	_ = cmd.MarkFlagRequired("longitude")
	_ = cmd.MarkFlagRequired("latitude")
	_ = cmd.MarkFlagRequired("address")

	return cmd
}

func newCmdLocationUnpin(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpin <device-id>",
		Short: "Unpin device location",
		Long:  "Remove the pinned location and restore automatic positioning (GPS/cell towers).",
		Example: `  # Unpin device location
  incloud device location unpin 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			payload := map[string]interface{}{"pinned": false}

			body, err := client.Put("/api/v1/devices/"+deviceID+"/location", payload)
			if err != nil {
				return err
			}

			locBody, err := extractLocation(body)
			if err != nil {
				return err
			}
			return formatOutput(cmd, f.IO, locBody)
		},
	}

	return cmd
}

func newCmdLocationRefresh(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh <device-id>",
		Short: "Refresh device location",
		Long:  "Trigger a location refresh using LBS (cell tower positioning).",
		Example: `  # Refresh device location
  incloud device location refresh 507f1f77bcf86cd799439011`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			body, err := client.Put("/api/v1/devices/"+deviceID+"/locations/refresh", nil)
			if err != nil {
				return err
			}

			locBody, err := extractLocation(body)
			if err != nil {
				return err
			}
			return formatOutput(cmd, f.IO, locBody)
		},
	}

	return cmd
}

// extractLocation extracts the location field from a device GET response.
// Input: {"result":{"location":{...},...}} → Output: {"result":{...}}
func extractLocation(body []byte) ([]byte, error) {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing device response: %w", err)
	}

	var device struct {
		Location json.RawMessage `json:"location"`
	}
	if err := json.Unmarshal(envelope.Result, &device); err != nil {
		return nil, fmt.Errorf("parsing device result: %w", err)
	}

	if device.Location == nil || string(device.Location) == "null" {
		return nil, fmt.Errorf("device has no location data")
	}

	return json.Marshal(map[string]json.RawMessage{"result": device.Location})
}
