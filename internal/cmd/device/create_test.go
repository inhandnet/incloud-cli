package device

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validSNResponseMAC = `{"result":{"product":"IR615","validatedField":"mac"}}`
const validSNResponseIMEI = `{"result":{"product":"IG902","validatedField":"imei"}}`
const defaultCreateResponse = `{"result":{"_id":"dev1","name":"test"}}`

// newCreateTestServer creates a test server that handles SN validation with validateResp
// and optionally a device create endpoint via createHandler.
// If createHandler is nil, create requests return 404.
func newCreateTestServer(t *testing.T, validateResp string, createHandler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/serialnumber/"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(validateResp))
		case createHandler != nil && r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices":
			createHandler(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestCreateDevice_FullFlow(t *testing.T) {
	var validateCalled, createCalled bool
	var createBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/serialnumber/"):
			validateCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(validSNResponseMAC))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/devices":
			createCalled = true
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result":{"_id":"dev123","name":"My Router"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "My Router", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !validateCalled {
		t.Error("expected SN validation API to be called")
	}
	if !createCalled {
		t.Error("expected device create API to be called")
	}
	if !strings.Contains(errBuf.String(), `Device "My Router" created.`) {
		t.Errorf("expected success message in stderr, got: %s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "dev123") {
		t.Errorf("expected device ID in stderr, got: %s", errBuf.String())
	}
}

func TestCreateDevice_RequiresMAC_NonTTY(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseMAC, nil)
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "ABCDEFGHIJKLMNO"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when MAC not provided in non-TTY")
	}
	if !strings.Contains(err.Error(), "--mac") {
		t.Errorf("expected --mac hint in error, got: %v", err)
	}
}

func TestCreateDevice_RequiresIMEI_NonTTY(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseIMEI, nil)
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "ABCDEFGHIJKLMNO"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when IMEI not provided in non-TTY")
	}
	if !strings.Contains(err.Error(), "--imei") {
		t.Errorf("expected --imei hint in error, got: %v", err)
	}
}

func TestCreateDevice_SNNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"resource_not_found"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "INVALIDSERIAL00"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid SN")
	}
	if !strings.Contains(err.Error(), "not recognized") {
		t.Errorf("expected 'not recognized' in error, got: %v", err)
	}
}

func TestCreateDevice_SNInvalidState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"invalid_state"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "OBSOLETESERIAL0"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for obsolete SN")
	}
	if !strings.Contains(err.Error(), "no longer supported") {
		t.Errorf("expected 'no longer supported' in error, got: %v", err)
	}
}

func TestCreateDevice_DuplicateName(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseMAC, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"resource_already_exists","ext":{"type":"name"}}`))
	})
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "duplicate", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected device name in error, got: %v", err)
	}
}

func TestCreateDevice_DuplicateSN(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseMAC, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"resource_already_exists","ext":{"type":"serialNumber"}}`))
	})
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for duplicate SN")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got: %v", err)
	}
}

func TestCreateDevice_MACInvalid(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseMAC, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"request_not_allowed","status":400,"message":"The device info is incorrect.","description":"MAC_INVALID"}`))
	})
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid MAC")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Errorf("expected 'does not match' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "AA:BB:CC:DD:EE:FF") {
		t.Errorf("expected MAC address in error, got: %v", err)
	}
}

func TestCreateDevice_AutoProduct(t *testing.T) {
	var createBody map[string]interface{}

	server := newCreateTestServer(t, validSNResponseMAC, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&createBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(defaultCreateResponse))
	})
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	product, ok := createBody["product"].(string)
	if !ok || product != "IR615" {
		t.Errorf("expected product IR615 in create body, got: %v", createBody["product"])
	}
}

func TestCreateDevice_SNUppercase(t *testing.T) {
	var validatePath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/serialnumber/") {
			validatePath = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(validSNResponseMAC))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(defaultCreateResponse))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "test", "--sn", "abcdefghijklmno", "--mac", "AA:BB:CC:DD:EE:FF"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(validatePath, "ABCDEFGHIJKLMNO") {
		t.Errorf("expected uppercase SN in validate path, got: %s", validatePath)
	}
}

func TestCreateDevice_SuccessMessage(t *testing.T) {
	server := newCreateTestServer(t, validSNResponseMAC, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"_id":"abc123","name":"Office Router"}}`))
	})
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	cmd := NewCmdCreate(f)
	cmd.SetArgs([]string{"--name", "Office Router", "--sn", "ABCDEFGHIJKLMNO", "--mac", "AA:BB:CC:DD:EE:FF"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errOutput := errBuf.String()
	if !strings.Contains(errOutput, `Device "Office Router" created.`) {
		t.Errorf("expected success message with device name, got: %s", errOutput)
	}
	if !strings.Contains(errOutput, "abc123") {
		t.Errorf("expected device ID in success message, got: %s", errOutput)
	}
}
