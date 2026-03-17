package device

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

func NewCmdUnassign(f *factory.Factory) *cobra.Command {
	var retainConfig bool

	cmd := &cobra.Command{
		Use:   "unassign <device-id>",
		Short: "Remove a device from its device group",
		Example: `  # Remove device from its group
  incloud device unassign 507f1f77bcf86cd799439011

  # Remove but retain group configuration
  incloud device unassign 507f1f77bcf86cd799439011 --retain-config`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID := args[0]

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			// GET device to find its oid and verify it belongs to a group
			dev, err := getDeviceInfo(client, deviceID)
			if err != nil {
				return err
			}
			if dev.groupName == "" {
				return fmt.Errorf("device %s is not in any group", deviceID)
			}

			body := map[string]interface{}{
				"items": []map[string]interface{}{
					{
						"deviceId": deviceID,
						"oid":      dev.oid,
					},
				},
			}
			if retainConfig {
				body["retainGroupConfig"] = true
			}

			_, err = client.Put("/api/v1/devices/move", body)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IO.ErrOut, "Device %s removed from its group.\n", deviceID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&retainConfig, "retain-config", false, "Retain group configuration after removal")

	return cmd
}

type deviceInfo struct {
	oid       string
	groupName string
}

func getDeviceInfo(client *api.APIClient, deviceID string) (*deviceInfo, error) {
	query := url.Values{}
	query.Set("fields", "oid,devicegroup")

	body, err := client.Get("/api/v1/devices/"+deviceID, query)
	if err != nil {
		return nil, fmt.Errorf("fetching device: %w", err)
	}

	var data struct {
		Result struct {
			Oid         string `json:"oid"`
			DeviceGroup struct {
				Name string `json:"name"`
			} `json:"devicegroup"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("parsing device response: %w", err)
	}
	if data.Result.Oid == "" {
		return nil, fmt.Errorf("device %s has no organization", deviceID)
	}
	return &deviceInfo{
		oid:       data.Result.Oid,
		groupName: data.Result.DeviceGroup.Name,
	}, nil
}
