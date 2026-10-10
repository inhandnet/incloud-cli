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

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

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
