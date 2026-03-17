package device

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newTestFactory(t *testing.T, host string) (*factory.Factory, *bytes.Buffer) {
	t.Helper()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {
				Host:  host,
				Token: "test-token",
			},
		},
	}
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	errBuf := &bytes.Buffer{}
	f := &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: errBuf,
		},
		ConfigPath: cfgPath,
	}
	return f, errBuf
}

func TestDeleteDevice_WithYesFlag(t *testing.T) {
	var gotMethod, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "--yes"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/api/v1/devices/abc123" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(errBuf.String(), "Device abc123 deleted.") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

func TestDeleteDevice_WithYesFlag_200Response(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"_id":"abc123","name":"test"}}`))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errBuf.String(), "Device abc123 deleted.") {
		t.Errorf("unexpected output: %s", errBuf.String())
	}
}

func TestDeleteDevice_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"device not found"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{"notfound", "--yes"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "device not found") {
		t.Errorf("expected error body in error message, got: %v", err)
	}
}

func TestDeleteDevice_NonTTYWithoutYes(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	// In is a strings.Reader (not a *os.File), so it's non-TTY
	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-TTY without --yes")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected --yes hint in error, got: %v", err)
	}
}

func TestDeleteDevice_ConfirmationAbort(t *testing.T) {
	// Use a pipe so In is an *os.File, but simulate non-"y" input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.WriteString("n\n")
	_ = w.Close()

	f, errBuf := newTestFactory(t, "https://example.com")
	f.IO.In = r

	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123"})
	// This will fail the isatty check since a pipe is not a terminal
	execErr := cmd.Execute()
	// Pipe is not a TTY, so it should error asking for --yes
	if execErr == nil {
		// If it somehow got through (shouldn't happen with pipe), check abort
		if strings.Contains(errBuf.String(), "Aborted") {
			return // acceptable
		}
		t.Fatal("expected error or abort")
	}
	if !strings.Contains(execErr.Error(), "--yes") {
		t.Errorf("expected --yes hint, got: %v", execErr)
	}
}

func TestDeleteDevice_RequiresExactlyOneArg(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	cmd := NewCmdDelete(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for no args")
	}

	cmd = NewCmdDelete(f)
	cmd.SetArgs([]string{"a", "b"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}

func TestDeleteDevice_AliasRm(t *testing.T) {
	cmd := NewCmdDelete(&factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	})
	found := false
	for _, a := range cmd.Aliases {
		if a == "rm" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'rm' alias")
	}
}
