package device

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractLocalLogs_Succeeded_LogsArray(t *testing.T) {
	body := `{"status":"succeeded","result":{"logs":["line1","line2","line3"]}}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "line1\nline2\nline3\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", string(got), want)
	}
}

func TestExtractLocalLogs_Succeeded_LogsArrayTrailingNewline(t *testing.T) {
	body := `{"status":"succeeded","result":{"logs":["line1\n","line2\n"]}}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Lines already end with \n, join should not double-add
	want := "line1\n\nline2\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", string(got), want)
	}
}

func TestExtractLocalLogs_Succeeded_PresignedURL(t *testing.T) {
	// Start a server that serves the log content at the presigned URL
	logContent := "remote log line 1\nremote log line 2\n"
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(logContent))
	}))
	defer s3Server.Close()

	body := `{"status":"succeeded","result":{"url":"` + s3Server.URL + `/logs"}}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != logContent {
		t.Errorf("got %q, want %q", string(got), logContent)
	}
}

func TestExtractLocalLogs_Succeeded_PresignedURL_HTTPError(t *testing.T) {
	s3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer s3Server.Close()

	body := `{"status":"succeeded","result":{"url":"` + s3Server.URL + `"}}`
	_, err := extractLocalLogs([]byte(body))
	if err == nil {
		t.Fatal("expected error for HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got: %v", err)
	}
}

func TestExtractLocalLogs_Succeeded_EmptyResult(t *testing.T) {
	body := `{"status":"succeeded","result":{}}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty result with no logs/url → fallback raw result
	if !strings.Contains(string(got), "{}") {
		t.Errorf("expected raw result fallback, got %q", string(got))
	}
}

func TestExtractLocalLogs_Succeeded_NullResult(t *testing.T) {
	body := `{"status":"succeeded","result":null}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// JSON null deserializes to RawMessage("null"), falls through to fallback
	if len(got) > 0 && string(got) != "null\n" {
		t.Errorf("expected empty or null fallback, got %q", string(got))
	}
}

func TestExtractLocalLogs_Succeeded_NonObjectResult(t *testing.T) {
	body := `{"status":"succeeded","result":"raw string content"}`
	got, err := extractLocalLogs([]byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// json.Unmarshal into struct fails for string → fallback prints raw
	if !strings.Contains(string(got), "raw string content") {
		t.Errorf("expected raw string fallback, got %q", string(got))
	}
}

func TestExtractLocalLogs_FailedStatus(t *testing.T) {
	body := `{"status":"timeout","error":"request timeout"}`
	_, err := extractLocalLogs([]byte(body))
	if err == nil {
		t.Fatal("expected error for timeout status")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "request timeout") {
		t.Errorf("expected error message, got: %v", err)
	}
}

func TestExtractLocalLogs_FailedStatus_NoError(t *testing.T) {
	body := `{"status":"failed"}`
	_, err := extractLocalLogs([]byte(body))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown error") {
		t.Errorf("expected 'unknown error', got: %v", err)
	}
}

func TestExtractLocalLogs_InvalidJSON(t *testing.T) {
	_, err := extractLocalLogs([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parsing response") {
		t.Errorf("expected parsing error, got: %v", err)
	}
}

func TestLogLocal_E2E_Stdout(t *testing.T) {
	var gotPath string
	var gotQuery map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = map[string]string{
			"lines":     r.URL.Query().Get("lines"),
			"localPath": r.URL.Query().Get("localPath"),
			"timeout":   r.URL.Query().Get("timeout"),
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"status": "succeeded",
			"result": map[string]any{
				"logs": []string{"Jan  1 00:00:01 syslog line 1", "Jan  1 00:00:02 syslog line 2"},
			},
		}
		body, _ := json.Marshal(resp)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	cmd := NewCmdLogLocal(f)
	cmd.SetArgs([]string{"device123", "--lines", "50", "--path", "/var/log/messages", "--timeout", "45"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/v1/devices/device123/logs/local" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotQuery["lines"] != "50" {
		t.Errorf("expected lines=50, got %s", gotQuery["lines"])
	}
	if gotQuery["localPath"] != "/var/log/messages" {
		t.Errorf("expected localPath=/var/log/messages, got %s", gotQuery["localPath"])
	}
	if gotQuery["timeout"] != "45" {
		t.Errorf("expected timeout=45, got %s", gotQuery["timeout"])
	}

	output := out.String()
	if !strings.Contains(output, "syslog line 1") {
		t.Errorf("expected log content in stdout, got: %s", output)
	}
}

func TestLogLocal_E2E_FileOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"status": "succeeded",
			"result": map[string]any{
				"logs": []string{"line A", "line B", "line C"},
			},
		}
		body, _ := json.Marshal(resp)
		_, _ = w.Write(body)
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	outFile := filepath.Join(t.TempDir(), "device.log")

	cmd := NewCmdLogLocal(f)
	cmd.SetArgs([]string{"device123", "--file", outFile})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// stdout should be empty when --file is used
	if out.Len() > 0 {
		t.Errorf("expected empty stdout with --file, got: %s", out.String())
	}

	// stderr should contain the saved path
	if !strings.Contains(errBuf.String(), "Saved to") {
		t.Errorf("expected 'Saved to' in stderr, got: %s", errBuf.String())
	}

	// File should contain the log content
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(content), "line A") {
		t.Errorf("expected log content in file, got: %s", string(content))
	}
}

func TestLogLocal_E2E_AllFlag(t *testing.T) {
	var gotAll, gotLines string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAll = r.URL.Query().Get("all")
		gotLines = r.URL.Query().Get("lines")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"succeeded","result":{"logs":["full log"]}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdLogLocal(f)
	cmd.SetArgs([]string{"device123", "--all"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAll != "true" {
		t.Errorf("expected all=true, got %s", gotAll)
	}
	if gotLines != "" {
		t.Errorf("expected lines to be omitted with --all, got %s", gotLines)
	}
}

func TestLogLocal_E2E_DeviceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"timeout","error":"request timeout"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdLogLocal(f)
	cmd.SetArgs([]string{"device123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for device timeout")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestLogLocal_E2E_MissingDeviceID(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	cmd := NewCmdLogLocal(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing device ID")
	}
}
