package device

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

// mqttLogBody mimics the cursor-paginated backend payload: entries ascending in time.
const mqttLogBody = `{"result":[` +
	`{"timestamp":"2026-08-04T00:00:00Z","logType":"publish"},` +
	`{"timestamp":"2026-08-07T00:00:00Z","logType":"publish"},` +
	`{"timestamp":"2026-08-10T17:00:00Z","logType":"publish"}` +
	`],"next":"cursor-token"}`

func runLogMqtt(t *testing.T, args ...string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mqttLogBody))
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	out := &bytes.Buffer{}
	f.IO.Out = out

	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	root.SetArgs(append([]string{"device", "log", "mqtt", "507f1f77bcf86cd799439011"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("device log mqtt %v: %v", args, err)
	}
	return out.String()
}

func firstLogTimestamp(t *testing.T, output string) string {
	t.Helper()

	var payload struct {
		Result []struct {
			Timestamp string `json:"timestamp"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json output is not the paginated shape (%v): %s", err, output)
	}
	if len(payload.Result) != 3 {
		t.Fatalf("expected three entries, got %d: %s", len(payload.Result), output)
	}
	return payload.Result[0].Timestamp
}

// TestLogMqttJSONHonorsOrder covers the second half of IM-3194: log mqtt reorders
// through the same table-only path signal list did, so its json callers were
// served the backend order regardless of --order.
func TestLogMqttJSONHonorsOrder(t *testing.T) {
	t.Run("desc puts the newest entry first", func(t *testing.T) {
		got := firstLogTimestamp(t, runLogMqtt(t, "-o", "json", "--order", "desc"))
		if got != "2026-08-10T17:00:00Z" {
			t.Errorf("expected the newest entry first, got %s", got)
		}
	})

	t.Run("asc keeps the backend order", func(t *testing.T) {
		got := firstLogTimestamp(t, runLogMqtt(t, "-o", "json", "--order", "asc"))
		if got != "2026-08-04T00:00:00Z" {
			t.Errorf("expected the oldest entry first, got %s", got)
		}
	})

	t.Run("pagination cursor survives reordering", func(t *testing.T) {
		out := runLogMqtt(t, "-o", "json", "--order", "desc")
		var payload struct {
			Next string `json:"next"`
		}
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if payload.Next != "cursor-token" {
			t.Errorf("expected the cursor to survive, got %q in %s", payload.Next, out)
		}
	})
}
