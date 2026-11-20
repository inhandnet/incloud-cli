package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/cmd/device"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func TestSyslogRedirectDefaultsToText(t *testing.T) {
	const deviceID = "507f1f77bcf86cd799439011"
	const line = "<30> Nov 19 10:48:02 conf_proc[1043]: done\n"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/"+deviceID+"/logs/download/syslog" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":2,"result":["<30> Nov 19 10:48:02 conf_proc[1043]: done\n"]}`))
	}))
	defer srv.Close()
	t.Setenv("INCLOUD_HOST", srv.URL)
	t.Setenv("INCLOUD_TOKEN", "env-token")

	var out, errOut bytes.Buffer
	f := &factory.Factory{
		IO:         &iostreams.IOStreams{Out: &out, ErrOut: &errOut},
		ConfigPath: filepath.Join(t.TempDir(), "config.yaml"),
	}
	root := NewCmdRoot(f)
	root.AddCommand(device.NewCmdDevice(f))
	root.SetArgs([]string{"device", "log", "syslog", deviceID,
		"--after", "2026-11-19T00:00:00Z", "--before", "2026-11-20T00:00:00Z", "--limit", "1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != line {
		t.Errorf("redirected stdout = %q, want %q", got, line)
	}
	if got := errOut.String(); !strings.Contains(got, "Showing lines 1-1 of 2") || !strings.Contains(got, "--page 2") {
		t.Errorf("missing truncation hint on stderr: %q", got)
	}
}

func newSuperAdminRoot(t *testing.T) (*cobra.Command, *bytes.Buffer, *atomic.Int32) {
	t.Helper()
	var meCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/users/me" {
			meCalls.Add(1)
			_, _ = w.Write([]byte(`{"result":{"roles":[{"name":"root","builtInRole":true}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":{}}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("INCLOUD_HOST", srv.URL)
	t.Setenv("INCLOUD_TOKEN", "env-token")

	out := &bytes.Buffer{}
	f := &factory.Factory{
		IO:         &iostreams.IOStreams{In: strings.NewReader(""), Out: out, ErrOut: &bytes.Buffer{}},
		ConfigPath: filepath.Join(t.TempDir(), "config.yaml"),
	}
	root := NewCmdRoot(f)
	root.SetOut(out)
	root.AddCommand(&cobra.Command{Use: "noop", RunE: func(*cobra.Command, []string) error { return nil }})
	SetupSuperAdminFlags(root, f)
	return root, out, &meCalls
}

func TestSuperAdminFlags_CheckedOnlyWhenRenderingHelp(t *testing.T) {
	root, _, meCalls := newSuperAdminRoot(t)
	t.Setenv("INCLOUD_SUDO", "") // restored after the --sudo flag sets it

	root.SetArgs([]string{"noop", "--sudo", "someone"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if n := meCalls.Load(); n != 0 {
		t.Errorf("users/me calls during command execution = %d, want 0", n)
	}
}

func TestSuperAdminFlags_HelpShowsSudoForSuperAdmin(t *testing.T) {
	root, out, meCalls := newSuperAdminRoot(t)

	root.SetArgs([]string{"noop", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if n := meCalls.Load(); n != 1 {
		t.Errorf("users/me calls = %d, want 1", n)
	}
	if !strings.Contains(out.String(), "--sudo") {
		t.Errorf("help output missing --sudo:\n%s", out.String())
	}
}
