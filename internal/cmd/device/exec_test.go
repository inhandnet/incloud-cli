package device

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExecMethod_SingleDevice(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"requestId":"abc","status":"succeeded"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"method", "device123", "myMethod", "--payload", `{"key":"val"}`, "--timeout", "15"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/api/v1/devices/device123/methods" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotBody["method"] != "myMethod" {
		t.Errorf("expected method=myMethod, got %v", gotBody["method"])
	}
	if gotBody["timeout"] != float64(15) {
		t.Errorf("expected timeout=15, got %v", gotBody["timeout"])
	}
	if !strings.Contains(out.String(), "succeeded") {
		t.Errorf("expected 'succeeded' in output, got: %s", out.String())
	}
}

func TestExecMethod_BulkDevices(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"method", "id1,id2,id3", "syncTime"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/bulk-invoke-methods" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	ids, ok := gotBody["deviceIds"].([]interface{})
	if !ok || len(ids) != 3 {
		t.Errorf("expected 3 deviceIds, got %v", gotBody["deviceIds"])
	}
	if !strings.Contains(errBuf.String(), "3 device(s)") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

func TestExecMethod_InvalidPayload(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"method", "device123", "myMethod", "--payload", "not-json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
	if !strings.Contains(err.Error(), "--payload") {
		t.Errorf("expected payload error hint, got: %v", err)
	}
}

func TestExecReboot_Confirmation(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	// In is strings.Reader (non-TTY), should require --yes

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"reboot", "device123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-TTY without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected --yes hint, got: %v", err)
	}
}

func TestExecReboot_WithYes(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"status":"succeeded"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"reboot", "device123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["method"] != "nezha_reboot" {
		t.Errorf("expected method=nezha_reboot, got %v", gotBody["method"])
	}
}

func TestExecRestoreDefaults_WithYes(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"status":"succeeded"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"restore-defaults", "device123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["method"] != "nezha_restore_to_defaults" {
		t.Errorf("expected method=nezha_restore_to_defaults, got %v", gotBody["method"])
	}
}

func TestExecPing_Stream(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}
	streamID := "stream-test-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/subscribe") {
			// SSE endpoint — matches real API format: sliding window of {index, content} items
			w.Header().Set("Content-Type", "text/event-stream")
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Fatal("expected http.Flusher")
			}
			// Event 1: lines 0-1 (sliding window)
			event1 := `{"status":"open","data":[` +
				`{"index":0,"content":"PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data."},` +
				`{"index":1,"content":"64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=3.45 ms"}]}`
			_, _ = w.Write([]byte("event: live\ndata: " + event1 + "\n\n"))
			flusher.Flush()
			// Event 2: lines 1-2 (overlapping window)
			event2 := `{"status":"closed","data":[` +
				`{"index":1,"content":"64 bytes from 8.8.8.8: icmp_seq=1 ttl=117 time=3.45 ms"},` +
				`{"index":2,"content":"--- 8.8.8.8 ping statistics ---"}]}`
			_, _ = w.Write([]byte("event: closed\ndata: " + event2 + "\n\n"))
			flusher.Flush()
			return
		}
		// POST diagnosis endpoint
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"_id":"diag123","streamId":"` + streamID + `"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"ping", "device123", "--host", "8.8.8.8"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/diagnosis/ping" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	output := out.String()
	if !strings.Contains(output, "PING 8.8.8.8") {
		t.Errorf("expected ping output, got: %s", output)
	}
	if !strings.Contains(output, "icmp_seq=1") {
		t.Errorf("expected icmp_seq=1 in output, got: %s", output)
	}
	if !strings.Contains(output, "ping statistics") {
		t.Errorf("expected statistics line from second event, got: %s", output)
	}
	// Verify deduplication: icmp_seq=1 should appear only once despite being in both events
	if strings.Count(output, "icmp_seq=1") != 1 {
		t.Errorf("expected icmp_seq=1 exactly once (dedup), got %d in: %s",
			strings.Count(output, "icmp_seq=1"), output)
	}
}

func TestExecPing_RequiresHost(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"ping", "device123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --host")
	}
}

func TestExecCapture_WaitForCompletion(t *testing.T) {
	// Speed up polling for tests
	oldInterval := capturePollInterval
	capturePollInterval = 10 * time.Millisecond
	defer func() { capturePollInterval = oldInterval }()

	var gotBody map[string]interface{}
	pollCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			_, _ = w.Write([]byte(`{"result":{"_id":"diag456","status":"RUNNING"}}`))
			return
		}
		// GET status polling
		pollCount++
		if pollCount < 2 {
			_, _ = w.Write([]byte(`{"result":{"_id":"diag456","status":"RUNNING"}}`))
		} else {
			_, _ = w.Write([]byte(`{"result":{"_id":"diag456","status":"FINISHED","fileUrl":"https://example.com/file.pcap"}}`))
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"capture", "device123", "--interface", "eth0", "--duration", "60"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotBody["interface"] != "eth0" {
		t.Errorf("expected interface=eth0, got %v", gotBody["interface"])
	}
	if gotBody["captureTime"] != float64(60) {
		t.Errorf("expected captureTime=60, got %v", gotBody["captureTime"])
	}
	if pollCount < 1 {
		t.Errorf("expected at least 1 poll, got %d", pollCount)
	}
	if !strings.Contains(out.String(), "FINISHED") {
		t.Errorf("expected FINISHED in output, got: %s", out.String())
	}
}

func TestExecCancel(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"cancel", "diag789"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/api/v1/diagnosis/diag789/cancel" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(errBuf.String(), "diag789 canceled") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

func TestExecCancel_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"cancel", "notexist"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404, got: %v", err)
	}
}

func TestExecSpeedtestHistory(t *testing.T) {
	var gotPath, gotPage, gotLimit, gotFrom string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotPage = r.URL.Query().Get("page")
		gotLimit = r.URL.Query().Get("limit")
		gotFrom = r.URL.Query().Get("from")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[],"total":0,"page":0,"limit":10}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"speedtest-history", "device123", "--page", "2", "--limit", "5", "--after", "2024-01-01T00:00:00Z"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/diagnosis/speed-test-histories" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotPage != "1" { // page 2 (1-based) → API page 1 (0-based)
		t.Errorf("expected page=1, got %s", gotPage)
	}
	if gotLimit != "5" {
		t.Errorf("expected limit=5, got %s", gotLimit)
	}
	if gotFrom != "2024-01-01T00:00:00Z" {
		t.Errorf("expected from=2024-01-01T00:00:00Z, got %s", gotFrom)
	}
}

func TestExecInterfaces(t *testing.T) {
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[{"name":"eth0","label":"Ethernet"}]}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"interfaces", "device123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/diagnosis/interfaces" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}
