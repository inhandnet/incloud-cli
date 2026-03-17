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

// runDiagnosis is a shared helper for POST diagnosis endpoints.
func runDiagnosis(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]interface{}) error {
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

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request body: %w", err)
	}

	reqURL := actx.Host + "/api/v1/devices/" + deviceID + "/diagnosis/" + tool
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
}

// getDiagnosisStatus is a shared helper for GET diagnosis status endpoints.
func getDiagnosisStatus(f *factory.Factory, cmd *cobra.Command, deviceID, tool string) error {
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

	reqURL := actx.Host + "/api/v1/devices/" + deviceID + "/diagnosis/" + tool
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return err
	}

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
}
