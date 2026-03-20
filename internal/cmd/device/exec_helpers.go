package device

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

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

// diagnosisStream holds the state returned by startDiagnosisStream.
type diagnosisStream struct {
	client   *api.APIClient
	streamID string
	ctx      context.Context
	cancel   context.CancelFunc
}

// startDiagnosisStream posts a diagnosis task and returns the stream state
// with a cancel function wired to Ctrl+C. Caller must defer ds.cancel().
func startDiagnosisStream(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]any) (diagnosisStream, error) {
	if output, _ := cmd.Flags().GetString("output"); output != "" {
		return diagnosisStream{}, fmt.Errorf("--output is not supported for streaming commands; output format is controlled by the device")
	}
	client, err := f.APIClient()
	if err != nil {
		return diagnosisStream{}, err
	}

	body := cleanDiagnosisParams(params)
	respBody, err := client.Post("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, body)
	if err != nil {
		return diagnosisStream{}, err
	}

	taskID := gjson.GetBytes(respBody, "result._id").String()
	streamID := gjson.GetBytes(respBody, "result.streamId").String()
	if streamID == "" {
		return diagnosisStream{}, fmt.Errorf("no streamId in response: %s", string(respBody))
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			cancel()
			if taskID != "" {
				_, _ = client.Put("/api/v1/diagnosis/"+taskID+"/cancel", nil)
			}
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()

	return diagnosisStream{client: client, streamID: streamID, ctx: ctx, cancel: cancel}, nil
}

// runDiagnosisStream starts a diagnosis task, subscribes to its SSE stream,
// and prints each result line to stdout in real time (append mode).
// On Ctrl+C, it cancels the task before exiting.
func runDiagnosisStream(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]any) error {
	ds, err := startDiagnosisStream(f, cmd, deviceID, tool, params)
	if err != nil {
		return err
	}
	defer ds.cancel()

	// Append mode: track the highest printed index to deduplicate overlapping
	// sliding windows across events.
	sseURL := ds.client.BaseURL() + "/api/v1/streams/" + ds.streamID + "/subscribe"
	maxPrinted := -1
	return api.StreamSSE(ds.ctx, ds.client.HTTPClient(), sseURL, func(event api.SSEEvent) {
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

// runDiagnosisStreamReplace starts a diagnosis task and streams output in
// replace mode: each SSE event contains the full current state, so previous
// output is cleared and reprinted. In TTY mode, ANSI escape codes are used
// to overwrite; in non-TTY mode, only the final state is printed.
func runDiagnosisStreamReplace(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, params map[string]any) error {
	ds, err := startDiagnosisStream(f, cmd, deviceID, tool, params)
	if err != nil {
		return err
	}
	defer ds.cancel()

	sseURL := ds.client.BaseURL() + "/api/v1/streams/" + ds.streamID + "/subscribe"
	isTTY := f.IO.IsStdoutTTY()
	prevLines := 0
	var lastContent []string

	err = api.StreamSSE(ds.ctx, ds.client.HTTPClient(), sseURL, func(event api.SSEEvent) {
		items := gjson.Get(event.Data, "data").Array()

		// Collect all content lines from this event
		var lines []string
		for _, item := range items {
			content := item.Get("content").String()
			if content != "" {
				lines = append(lines, content)
			}
		}
		if len(lines) == 0 {
			return
		}

		lastContent = lines

		if isTTY {
			// Clear previous output by moving cursor up and clearing lines
			if prevLines > 0 {
				fmt.Fprintf(f.IO.Out, "\033[%dA\033[J", prevLines)
			}
			printed := 0
			for _, line := range lines {
				fmt.Fprintln(f.IO.Out, line)
				// Count actual terminal lines (content may contain embedded newlines)
				printed += 1 + strings.Count(line, "\n")
			}
			prevLines = printed
		}
	})

	// Non-TTY: print only the final state once the stream ends
	if !isTTY && len(lastContent) > 0 {
		for _, line := range lastContent {
			fmt.Fprintln(f.IO.Out, line)
		}
	}
	return err
}

// getDiagnosisStatus is a shared helper for GET diagnosis status endpoints.
// An optional url.Values can be passed to add query parameters.
func getDiagnosisStatus(f *factory.Factory, cmd *cobra.Command, deviceID, tool string, query ...url.Values) error {
	client, err := f.APIClient()
	if err != nil {
		return err
	}

	q := url.Values{}
	if len(query) > 0 && query[0] != nil {
		q = query[0]
	}

	respBody, err := client.Get("/api/v1/devices/"+deviceID+"/diagnosis/"+tool, q)
	if err != nil {
		return err
	}

	return formatOutput(cmd, f.IO, respBody, nil)
}
