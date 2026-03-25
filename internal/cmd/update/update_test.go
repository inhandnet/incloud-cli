package update

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newTestIO() (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &iostreams.IOStreams{
		In:     strings.NewReader(""),
		Out:    out,
		ErrOut: errOut,
	}, out, errOut
}

func TestPrintCheckResult_JSON_UpdateAvailable(t *testing.T) {
	io, out, _ := newTestIO()

	err := printCheckResult(io, "json", "v0.1.0", "v0.2.0", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result checkResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.Current != "v0.1.0" {
		t.Errorf("expected current v0.1.0, got %s", result.Current)
	}
	if result.Latest != "v0.2.0" {
		t.Errorf("expected latest v0.2.0, got %s", result.Latest)
	}
	if !result.UpdateAvailable {
		t.Error("expected update_available to be true")
	}
}

func TestPrintCheckResult_JSON_AlreadyUpToDate(t *testing.T) {
	io, out, _ := newTestIO()

	err := printCheckResult(io, "json", "v0.2.0", "v0.2.0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result checkResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.UpdateAvailable {
		t.Error("expected update_available to be false")
	}
}

func TestPrintCheckResult_Text_UpdateAvailable(t *testing.T) {
	io, _, errOut := newTestIO()

	err := printCheckResult(io, "", "v0.1.0", "v0.2.0", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := errOut.String()
	if !strings.Contains(output, "v0.2.0") {
		t.Errorf("expected version in output, got: %s", output)
	}
	if !strings.Contains(output, "incloud update") {
		t.Errorf("expected update hint in output, got: %s", output)
	}
}

func TestPrintCheckResult_Text_AlreadyUpToDate(t *testing.T) {
	io, _, errOut := newTestIO()

	err := printCheckResult(io, "", "v0.2.0", "v0.2.0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(errOut.String(), "Already up to date") {
		t.Errorf("expected 'Already up to date', got: %s", errOut.String())
	}
}

func TestConfirmUpdate_SkipConfirm(t *testing.T) {
	io, _, _ := newTestIO()
	cancelled := confirmUpdate(io, true)
	if cancelled {
		t.Error("expected not cancelled when skipConfirm=true")
	}
}

func TestConfirmUpdate_NonTTY(t *testing.T) {
	// Non-TTY IOStreams (default from newTestIO) should skip confirmation
	io, _, _ := newTestIO()
	cancelled := confirmUpdate(io, false)
	if cancelled {
		t.Error("expected not cancelled for non-TTY")
	}
}

func TestDevBuildGuard(t *testing.T) {
	// build.Version defaults to "dev" in test builds
	f := &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdUpdate(f)
	cmd.SetArgs([]string{"--check"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if !strings.Contains(err.Error(), "development build") {
		t.Errorf("expected dev build error, got: %v", err)
	}
}

func TestUpdateCommand_Flags(t *testing.T) {
	f := &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}

	cmd := NewCmdUpdate(f)

	flags := []string{"check", "version", "yes", "output"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}

	// Check short flags
	if cmd.Flags().ShorthandLookup("y") == nil {
		t.Error("expected short flag -y")
	}
	if cmd.Flags().ShorthandLookup("o") == nil {
		t.Error("expected short flag -o")
	}
}
