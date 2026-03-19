package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  string
	ID    string
}

// StreamSSE opens an SSE connection and calls onEvent for each event received.
// It blocks until the stream ends, the context is canceled, or an error occurs.
func StreamSSE(ctx context.Context, client *http.Client, url string, onEvent func(SSEEvent)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SSE HTTP %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1 MB max line
	var event SSEEvent

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = event boundary; dispatch if we have data
			if event.Data != "" {
				onEvent(event)
			}
			event = SSEEvent{}
			continue
		}

		if strings.HasPrefix(line, ":") {
			// Comment line, skip
			continue
		}

		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += value
		case "event":
			event.Event = value
		case "id":
			event.ID = value
		}
	}

	// Dispatch any remaining event
	if event.Data != "" {
		onEvent(event)
	}

	if err := scanner.Err(); err != nil {
		// Context cancellation is not an error for the caller
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("SSE read error: %w", err)
	}

	return nil
}
