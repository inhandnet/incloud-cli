package device

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"

	"github.com/inhandnet/incloud-cli/internal/api"
	"github.com/inhandnet/incloud-cli/internal/factory"
)

// cleanDiagnosisParams removes zero-value entries from a params map.
func cleanDiagnosisParams(params map[string]any) map[string]any {
	body := make(map[string]any)
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
	return body
}

// runDiagnosis is a shared helper for POST diagnosis endpoints.
func runDiagnosis(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]any) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := cleanDiagnosisParams(params)

	respBody, err := client.Post("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, body)
	if err != nil {
		return err
	}

	return formatOutput(cmd, f.IO, respBody, nil)
}

// runDiagnosisStream starts a diagnosis task, subscribes to its SSE stream,
// and prints each result line to stdout in real time. On Ctrl+C, it cancels
// the task before exiting.
func runDiagnosisStream(f *factory.Factory, deviceID, tool string, params map[string]any) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := cleanDiagnosisParams(params)

	// 1. POST to start the diagnosis task
	respBody, err := client.Post("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, body)
	if err != nil {
		return err
	}

	taskID := gjson.GetBytes(respBody, "result._id").String()
	streamID := gjson.GetBytes(respBody, "result.streamId").String()
	if streamID == "" {
		return fmt.Errorf("no streamId in response: %s", string(respBody))
	}

	// 2. Set up Ctrl+C to cancel the task
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)
	go func() {
		select {
		case <-sigCh:
			cancel()
			if taskID != "" {
				// Best-effort cancel
				_, _ = client.Put("/api/v1/diagnosis/"+taskID+"/cancel", nil)
			}
		case <-ctx.Done():
		}
	}()

	// 3. Subscribe to SSE stream and print results.
	// The SSE events contain a sliding window of result lines in data[].
	// Each line has an index sorted ascending; we track the highest printed
	// index to deduplicate overlapping windows across events.
	sseURL := client.BaseURL() + "/api/v1/streams/" + streamID + "/subscribe"
	maxPrinted := -1
	return api.StreamSSE(ctx, client.HTTPClient(), sseURL, func(event api.SSEEvent) {
		items := gjson.Get(event.Data, "data").Array()
		for _, item := range items {
			idx := int(item.Get("index").Int())
			if idx <= maxPrinted {
				continue
			}
			content := item.Get("content").String()
			if content != "" {
				fmt.Fprintln(f.IO.Out, content)
			}
			maxPrinted = idx
		}
	})
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
