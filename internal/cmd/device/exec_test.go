package device

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestExecPing(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"_id":"diag123","status":"RUNNING"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"ping", "device123", "--host", "8.8.8.8", "--count", "5"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/diagnosis/ping" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotBody["host"] != "8.8.8.8" {
		t.Errorf("expected host=8.8.8.8, got %v", gotBody["host"])
	}
	if gotBody["pingCount"] != float64(5) {
		t.Errorf("expected pingCount=5, got %v", gotBody["pingCount"])
	}
	// interface should be omitted when empty
	if _, ok := gotBody["interface"]; ok {
		t.Errorf("expected interface to be omitted, got %v", gotBody["interface"])
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

func TestExecCapture(t *testing.T) {
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"_id":"diag456","status":"RUNNING"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

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
}

func TestExecCaptureStatus(t *testing.T) {
	var gotMethod, gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"status":"FINISHED","fileUrl":"https://example.com/file.pcap"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"capture-status", "device123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/api/v1/devices/device123/diagnosis/capture" {
		t.Errorf("unexpected path: %s", gotPath)
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
	var gotPath, gotPage, gotSize, gotFrom string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotPage = r.URL.Query().Get("page")
		gotSize = r.URL.Query().Get("size")
		gotFrom = r.URL.Query().Get("from")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[],"total":0,"page":0,"limit":10}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdExec(f)
	cmd.SetArgs([]string{"speedtest-history", "device123", "--page", "2", "--limit", "5", "--after", "2024-01-01"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/diagnosis/speed-test-histories" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotPage != "1" { // page 2 (1-based) → API page 1 (0-based)
		t.Errorf("expected page=1, got %s", gotPage)
	}
	if gotSize != "5" {
		t.Errorf("expected size=5, got %s", gotSize)
	}
	if gotFrom != "2024-01-01" {
		t.Errorf("expected from=2024-01-01, got %s", gotFrom)
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
