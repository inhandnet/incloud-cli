package device

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// signalBody mimics the backend payload: samples ascending in time, the order
// the downsampled Flux query returns them in.
const signalBody = `{"result":{"series":[{"fields":["time","rsrp"],"data":[` +
	`["2026-08-04T00:00:00Z",-104],` +
	`["2026-08-07T00:00:00Z",-95],` +
	`["2026-08-10T17:00:00Z",-82]]}]}}`

func newSignalRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	return root
}

// runSignalList runs the command against a stub backend and returns what it printed.
func runSignalList(t *testing.T, args ...string) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(signalBody))
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	out := &bytes.Buffer{}
	f.IO.Out = out

	root := newSignalRoot(f)
	root.SetArgs(append([]string{"device", "signal", "list", "507f1f77bcf86cd799439011"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("device signal list %v: %v", args, err)
	}
	return out.String()
}

// samples reads back the series a structured caller would parse.
func samples(t *testing.T, output string) [][]interface{} {
	t.Helper()

	var payload struct {
		Series []struct {
			Fields []string        `json:"fields"`
			Data   [][]interface{} `json:"data"`
		} `json:"series"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("json output is not series-shaped (%v): %s", err, output)
	}
	if len(payload.Series) != 1 {
		t.Fatalf("expected one series, got %d: %s", len(payload.Series), output)
	}
	if len(payload.Series[0].Data) != 3 {
		t.Fatalf("expected three samples, got %d: %s", len(payload.Series[0].Data), output)
	}
	return payload.Series[0].Data
}

func firstTimestamp(t *testing.T, output string) string {
	t.Helper()

	ts, ok := samples(t, output)[0][0].(string)
	if !ok {
		t.Fatalf("first sample has no timestamp: %s", output)
	}
	return ts
}

// TestSignalListJSONHonorsOrder is the regression for IM-3194: --order used to
// apply to table rendering only, so json callers taking data[0] as the latest
// reading silently got the oldest sample in the window instead.
func TestSignalListJSONHonorsOrder(t *testing.T) {
	t.Run("desc puts the newest sample first", func(t *testing.T) {
		got := firstTimestamp(t, runSignalList(t, "-o", "json", "--order", "desc"))
		if got != "2026-08-10T17:00:00Z" {
			t.Errorf("expected the newest sample first, got %s", got)
		}
	})

	t.Run("default order is desc, as --help declares", func(t *testing.T) {
		got := firstTimestamp(t, runSignalList(t, "-o", "json"))
		if got != "2026-08-10T17:00:00Z" {
			t.Errorf("expected the newest sample first by default, got %s", got)
		}
	})

	t.Run("asc keeps the backend order", func(t *testing.T) {
		got := firstTimestamp(t, runSignalList(t, "-o", "json", "--order", "asc"))
		if got != "2026-08-04T00:00:00Z" {
			t.Errorf("expected the oldest sample first, got %s", got)
		}
	})

	t.Run("reordering keeps samples paired with their readings", func(t *testing.T) {
		rows := samples(t, runSignalList(t, "-o", "json", "--order", "desc"))
		if rows[0][1] != float64(-82) {
			t.Errorf("expected the newest sample to keep rsrp -82, got %v", rows[0][1])
		}
		if rows[2][1] != float64(-104) {
			t.Errorf("expected the oldest sample to keep rsrp -104, got %v", rows[2][1])
		}
	})
}
