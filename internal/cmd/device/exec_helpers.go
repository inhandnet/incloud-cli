package device

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// runDiagnosis is a shared helper for POST diagnosis endpoints.
func runDiagnosis(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]interface{}) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// Remove zero-value params
	body := make(map[string]interface{})
	for k, v := range params {
		switch val := v.(type) {
		case string:
			if val != "" {
				body[k] = val
			}
		case int:
			if val != 0 {
				body[k] = val
			}
		case []string:
			if len(val) > 0 {
				body[k] = val
			}
		default:
			if val != nil {
				body[k] = val
			}
		}
	}

	respBody, err := client.Post("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, body)
	if err != nil {
		return err
	}

	return formatOutput(cmd, f.IO, respBody, nil)
}

// getDiagnosisStatus is a shared helper for GET diagnosis status endpoints.
func getDiagnosisStatus(f *factory.Factory, cmd *cobra.Command, deviceID, tool string) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	respBody, err := client.Get("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, url.Values{})
	if err != nil {
		return err
	}

	return formatOutput(cmd, f.IO, respBody, nil)
}
