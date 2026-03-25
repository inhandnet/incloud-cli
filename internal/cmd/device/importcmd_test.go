package device

import (
	"encoding/csv"
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/inhandnet/incloud-cli/internal/api"
)

func TestCsvToXLSX(t *testing.T) {
	csvPath := filepath.Join(t.TempDir(), "test.csv")
	if err := os.WriteFile(csvPath, []byte("name,serialNumber,mac,imei\nDev1,SN001,AA:BB:CC:DD:EE:01,\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	xlsxPath, err := csvToXLSX(csvPath)
	if err != nil {
		t.Fatalf("csvToXLSX failed: %v", err)
	}
	defer func() { _ = os.Remove(xlsxPath) }()

	if filepath.Ext(xlsxPath) != ".xlsx" {
		t.Errorf("expected .xlsx extension, got %s", filepath.Ext(xlsxPath))
	}

	// Verify content
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		t.Fatalf("opening xlsx: %v", err)
	}
	defer func() { _ = f.Close() }()

	val, err := f.GetCellValue("Sheet1", "A1")
	if err != nil {
		t.Fatalf("reading A1: %v", err)
	}
	if val != "name" {
		t.Errorf("A1: expected 'name', got %q", val)
	}

	val, err = f.GetCellValue("Sheet1", "B2")
	if err != nil {
		t.Fatalf("reading B2: %v", err)
	}
	if val != "SN001" {
		t.Errorf("B2: expected 'SN001', got %q", val)
	}
}

func TestCsvToXLSX_HeaderOnly(t *testing.T) {
	csvPath := filepath.Join(t.TempDir(), "header_only.csv")
	if err := os.WriteFile(csvPath, []byte("name,serialNumber\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := csvToXLSX(csvPath)
	if err == nil {
		t.Fatal("expected error for header-only CSV")
	}
	if !strings.Contains(err.Error(), "at least one data row") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCsvToXLSX_FileNotFound(t *testing.T) {
	_, err := csvToXLSX("/nonexistent/file.csv")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestImport_UnsupportedFormat(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{"data.json", "-y"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported file format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestImport_NonTTYWithoutYes(t *testing.T) {
	// Server that handles upload + detail
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"job123"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/detail"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":{"_id":"job123","fileName":"test.xlsx","total":2,"status":"init","rate":0}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	// In is a strings.Reader (not *os.File), so non-TTY

	xlsxPath := createTestXLSX(t)
	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-TTY without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected --yes hint, got: %v", err)
	}
}

func TestImport_UploadAndConfirmSuccess(t *testing.T) {
	var (
		uploadReceived  atomic.Bool
		confirmReceived atomic.Bool
		detailCalls     atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			// Verify multipart upload
			ct := r.Header.Get("Content-Type")
			mt, _, err := mime.ParseMediaType(ct)
			if err != nil || mt != "multipart/form-data" {
				t.Errorf("expected multipart/form-data, got %s", ct)
			}
			uploadReceived.Store(true)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"job456"}`))

		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/job456/detail":
			n := detailCalls.Add(1)
			job := map[string]interface{}{
				"_id":      "job456",
				"fileName": "test.xlsx",
				"total":    2,
				"rate":     0.0,
				"status":   "init",
			}
			// First call: init (waitForValidation will see this)
			// Second call after confirm: importing
			// Third call: success
			switch n {
			case 1:
				job["status"] = "init"
			case 2:
				job["status"] = "importing"
				job["rate"] = 0.5
			default:
				job["status"] = "success"
				job["successNo"] = 2
				job["rate"] = 1.0
			}
			resp := map[string]interface{}{"result": job}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports/job456":
			confirmReceived.Store(true)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"job456"}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	xlsxPath := createTestXLSX(t)
	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uploadReceived.Load() {
		t.Error("upload was not received")
	}
	if !confirmReceived.Load() {
		t.Error("confirm was not received")
	}
	if !strings.Contains(errBuf.String(), "imported successfully") {
		t.Errorf("expected success message, got: %s", errBuf.String())
	}
}

func TestImport_UploadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"File type error"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	xlsxPath := createTestXLSX(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for upload failure")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected HTTP 400, got: %v", err)
	}
}

func TestImport_ValidationFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"job789"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/job789/detail":
			resp := `{"result":{"_id":"job789","fileName":"test.xlsx","total":2,"status":"check_fail","result":{"SERIAL_ILLEGAL":[2,3]},"rate":1}}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	xlsxPath := createTestXLSX(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for validation failure")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected validation error, got: %v", err)
	}
	if !strings.Contains(errBuf.String(), "SERIAL_ILLEGAL") {
		t.Errorf("expected error details in output, got: %s", errBuf.String())
	}
}

func TestImport_ImportWithFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobfail"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/jobfail/detail":
			resp := `{"result":{"_id":"jobfail","fileName":"test.xlsx","total":3,"successNo":1,"failNo":2,"status":"failed","result":{"SERIAL_ILLEGAL":[2,3]},"rate":1}}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports/jobfail":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobfail"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	xlsxPath := createTestXLSX(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for import with failures")
	}
	if !strings.Contains(err.Error(), "2 failure(s)") {
		t.Errorf("expected failure count, got: %v", err)
	}
	if !strings.Contains(errBuf.String(), "1 succeeded, 2 failed") {
		t.Errorf("expected result summary, got: %s", errBuf.String())
	}
}

func TestImport_NoWait(t *testing.T) {
	var confirmReceived atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobnw"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/jobnw/detail":
			resp := `{"result":{"_id":"jobnw","fileName":"test.xlsx","total":2,"status":"init","rate":0}}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports/jobnw":
			confirmReceived.Store(true)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobnw"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	xlsxPath := createTestXLSX(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y", "--no-wait"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !confirmReceived.Load() {
		t.Error("confirm was not called")
	}
	if !strings.Contains(errBuf.String(), "Import job jobnw started") {
		t.Errorf("expected no-wait message, got: %s", errBuf.String())
	}
}

func TestImport_WithGroup(t *testing.T) {
	var receivedGroupID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			if err := r.ParseMultipartForm(10 << 20); err == nil {
				receivedGroupID = r.FormValue("groupId")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobgrp"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/jobgrp/detail":
			resp := `{"result":{"_id":"jobgrp","fileName":"test.xlsx","total":2,"successNo":2,"failNo":0,"status":"success","rate":1}}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports/jobgrp":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobgrp"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	xlsxPath := createTestXLSX(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{xlsxPath, "-y", "--group", "507f1f77bcf86cd799439011"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedGroupID != "507f1f77bcf86cd799439011" {
		t.Errorf("expected groupId=507f1f77bcf86cd799439011 in form, got %q", receivedGroupID)
	}
}

func TestImport_CSVUpload(t *testing.T) {
	var receivedFileName string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports":
			// Parse multipart to get filename
			if err := r.ParseMultipartForm(10 << 20); err == nil {
				if fh := r.MultipartForm.File["file"]; len(fh) > 0 {
					receivedFileName = fh[0].Filename
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobcsv"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/devices/imports/jobcsv/detail":
			resp := `{"result":{"_id":"jobcsv","fileName":"test.xlsx","total":1,"successNo":1,"failNo":0,"status":"success","rate":1}}`
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(resp))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices/imports/jobcsv":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":"jobcsv"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	csvPath := createTestCSV(t)

	cmd := NewCmdImport(f)
	cmd.SetArgs([]string{csvPath, "-y"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have been converted to xlsx before upload
	if !strings.HasSuffix(receivedFileName, ".xlsx") {
		t.Errorf("expected .xlsx upload, got filename: %s", receivedFileName)
	}
}

func TestImport_WaitForValidation_CheckingThenReady(t *testing.T) {
	var detailCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := detailCalls.Add(1)
		job := map[string]interface{}{
			"_id":      "jobchk",
			"fileName": "test.xlsx",
			"total":    2,
			"rate":     0.0,
		}
		if n <= 2 {
			job["status"] = "checking"
		} else {
			// After checking completes, status becomes "init" (ready to confirm)
			job["status"] = "init"
		}
		resp := map[string]interface{}{"result": job}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	job, err := waitForValidation(api.NewAPIClient(server.URL, server.Client().Transport), "jobchk")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "init" {
		t.Errorf("expected status 'init', got %q", job.Status)
	}
	if got := detailCalls.Load(); got < 3 {
		t.Errorf("expected at least 3 detail calls (2 checking + 1 init), got %d", got)
	}
}

func TestImport_WaitForValidation_CheckFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{"result":{"_id":"jobchkf","fileName":"test.xlsx","total":2,"status":"check_fail","result":{"SERIAL_ILLEGAL":[2,3]},"rate":1}}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	job, err := waitForValidation(api.NewAPIClient(server.URL, server.Client().Transport), "jobchkf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "check_fail" {
		t.Errorf("expected status 'check_fail', got %q", job.Status)
	}
}

func TestImport_WaitForValidation_ImmediateInit(t *testing.T) {
	var detailCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detailCalls.Add(1)
		resp := `{"result":{"_id":"jobinit","fileName":"test.xlsx","total":2,"status":"init","rate":0}}`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	job, err := waitForValidation(api.NewAPIClient(server.URL, server.Client().Transport), "jobinit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != "init" {
		t.Errorf("expected status 'init', got %q", job.Status)
	}
	// Should return immediately on first call (init is not "checking")
	if got := detailCalls.Load(); got != 1 {
		t.Errorf("expected 1 detail call, got %d", got)
	}
}

func TestIsTerminalStatus(t *testing.T) {
	tests := []struct {
		status   string
		terminal bool
	}{
		{"success", true},
		{"failed", true},
		{"check_fail", true},
		{"cancel", true},
		{"init", false},
		{"checking", false},
		{"waiting", false},
		{"importing", false},
	}
	for _, tt := range tests {
		if got := isTerminalStatus(tt.status); got != tt.terminal {
			t.Errorf("isTerminalStatus(%q) = %v, want %v", tt.status, got, tt.terminal)
		}
	}
}

func TestShowImportResult_Success(t *testing.T) {
	f, errBuf := newTestFactory(t, "https://example.com")
	job := &importJob{Status: "success", SuccessNo: 5, Total: 5}
	err := showImportResult(f, job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "5/5 device(s) imported successfully") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

func TestShowImportResult_PartialFailure(t *testing.T) {
	f, errBuf := newTestFactory(t, "https://example.com")
	job := &importJob{
		Status:    "failed",
		Total:     5,
		SuccessNo: 3,
		FailNo:    2,
		Result:    map[string][]int{"SERIAL_ILLEGAL": {2, 4}},
	}
	err := showImportResult(f, job)
	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if !strings.Contains(errBuf.String(), "3 succeeded, 2 failed") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "SERIAL_ILLEGAL") {
		t.Errorf("expected error details: %s", errBuf.String())
	}
}

func TestShowImportResult_Cancelled(t *testing.T) {
	f, errBuf := newTestFactory(t, "https://example.com")
	job := &importJob{Status: "cancel", Total: 5}
	err := showImportResult(f, job)
	if err == nil {
		t.Fatal("expected error for cancelled import")
	}
	if !strings.Contains(errBuf.String(), "cancelled") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

// createTestXLSX creates a minimal valid XLSX file for testing.
func createTestXLSX(t *testing.T) string {
	t.Helper()
	xlsx := excelize.NewFile()
	defer func() { _ = xlsx.Close() }()

	sheet := "Sheet1"
	_ = xlsx.SetCellValue(sheet, "A1", "name")
	_ = xlsx.SetCellValue(sheet, "B1", "serialNumber")
	_ = xlsx.SetCellValue(sheet, "C1", "mac")
	_ = xlsx.SetCellValue(sheet, "D1", "imei")
	_ = xlsx.SetCellValue(sheet, "A2", "TestDevice")
	_ = xlsx.SetCellValue(sheet, "B2", "TEST001")

	path := filepath.Join(t.TempDir(), "test.xlsx")
	if err := xlsx.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	return path
}

// createTestCSV creates a minimal valid CSV file for testing.
func createTestCSV(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.csv")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"name", "serialNumber", "mac", "imei"})
	_ = w.Write([]string{"TestDevice", "TEST001", "AA:BB:CC:DD:EE:01", ""})
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	return path
}
