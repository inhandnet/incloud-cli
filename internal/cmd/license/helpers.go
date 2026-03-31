package license

import (
	"encoding/json"
	"fmt"
)

type licenseState struct {
	Status   string
	DeviceID string
}

func parseLicenseState(body []byte) (licenseState, error) {
	var resp struct {
		Result struct {
			Status   string `json:"status"`
			DeviceID string `json:"deviceId"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return licenseState{}, fmt.Errorf("failed to parse license data: %w", err)
	}
	return licenseState{
		Status:   resp.Result.Status,
		DeviceID: resp.Result.DeviceID,
	}, nil
}

type deviceLicenseState struct {
	ID     string
	Status string
}

func parseDeviceLicense(body []byte) (deviceLicenseState, error) {
	var resp struct {
		Result struct {
			License struct {
				ID     string `json:"_id"`
				Status string `json:"status"`
			} `json:"license"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return deviceLicenseState{}, fmt.Errorf("failed to parse device data: %w", err)
	}
	return deviceLicenseState{
		ID:     resp.Result.License.ID,
		Status: resp.Result.License.Status,
	}, nil
}
