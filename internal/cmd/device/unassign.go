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

			// GET device to find its oid and verify it belongs to a group
			dev, err := getDeviceInfo(client, actx.Host, deviceID)
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

			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			reqURL := actx.Host + "/api/v1/devices/move"
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

			if resp.StatusCode >= 400 {
				fmt.Fprintln(f.IO.ErrOut, string(respBody))
				return fmt.Errorf("HTTP %d", resp.StatusCode)
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

func getDeviceInfo(client *http.Client, host, deviceID string) (*deviceInfo, error) {
	reqURL := host + "/api/v1/devices/" + deviceID + "?fields=oid,devicegroup"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching device: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading device response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fetching device: HTTP %d: %s", resp.StatusCode, string(body))
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
