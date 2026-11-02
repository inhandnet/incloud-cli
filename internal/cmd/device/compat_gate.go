package device

import (
	"encoding/json"
	"fmt"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// compatibilityVerdict is one entry of a bulk-validate result.
type compatibilityVerdict struct {
	Support         bool   `json:"support"`
	MinVersion      string `json:"minVersion"`
	FirmwareUnknown bool   `json:"firmwareUnknown"`
}

// checkCompatibility asks the platform whether the device supports compatID.
// It returns nil when the answer is unknown (request failed, or the key is absent
// from the result); callers then proceed as if supported, so a missing permission
// or a key the platform does not define never hides data.
func checkCompatibility(client *api.APIClient, deviceID, compatID string) *compatibilityVerdict {
	body, err := client.Post("/api/v1/product-compatibilities/bulk-validate", map[string]interface{}{
		"ids":             []string{deviceID},
		"compatibilities": []string{compatID},
		"type":            "DEVICE",
	})
	if err != nil {
		return nil
	}
	var resp struct {
		Result []struct {
			Compatibilities map[string]compatibilityVerdict `json:"compatibilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || len(resp.Result) == 0 {
		return nil
	}
	v, ok := resp.Result[0].Compatibilities[compatID]
	if !ok {
		return nil
	}
	return &v
}

// unsupportedMessage mirrors the portal's SupportAlert wording.
func unsupportedMessage(v *compatibilityVerdict) string {
	switch {
	case v.FirmwareUnknown:
		return "The device has not been online yet, so its firmware version is unknown and support for this feature cannot be determined."
	case v.MinVersion != "":
		return fmt.Sprintf("The current firmware of this device does not support this feature, please upgrade its version to %s or above.", v.MinVersion)
	default:
		return "This product does not support this feature."
	}
}

// writeUnsupported prints the notice returned instead of data when the device does
// not support compatID.
func writeUnsupported(io *iostreams.IOStreams, output, deviceID, compatID string, v *compatibilityVerdict) error {
	msg := unsupportedMessage(v)
	if output == "table" && io.JQExpr == "" {
		fmt.Fprintln(io.Out, msg)
		return nil
	}
	notice := map[string]interface{}{
		"deviceId":      deviceID,
		"compatibility": compatID,
		"supported":     false,
		"message":       msg,
	}
	if v.MinVersion != "" {
		notice["minVersion"] = v.MinVersion
	}
	body, err := json.Marshal(notice)
	if err != nil {
		return err
	}
	return iostreams.FormatOutput(body, io, output)
}
