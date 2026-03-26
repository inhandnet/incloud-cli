package update

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/go-selfupdate"

	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// mockSource implements selfupdate.Source for testing fallback behavior.
type mockSource struct {
	releases    []selfupdate.SourceRelease
	listErr     error
	downloadRC  io.ReadCloser
	downloadErr error
}

func (m *mockSource) ListReleases(_ context.Context, _ selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.releases, nil
}

func (m *mockSource) DownloadReleaseAsset(_ context.Context, _ *selfupdate.Release, _ int64) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return m.downloadRC, nil
}

func newTestIO() (streams *iostreams.IOStreams, stdout, stderr *bytes.Buffer) {
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	streams = &iostreams.IOStreams{
		In:     strings.NewReader(""),
		Out:    stdout,
		ErrOut: stderr,
	}
	return
}

func TestPrintCheckResult_JSON_UpdateAvailable(t *testing.T) {
	ios, out, _ := newTestIO()

	err := printCheckResult(ios, "json", "v0.1.0", "v0.2.0", true)
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
	ios, out, _ := newTestIO()

	err := printCheckResult(ios, "json", "v0.2.0", "v0.2.0", false)
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
	ios, _, errOut := newTestIO()

	err := printCheckResult(ios, "", "v0.1.0", "v0.2.0", true)
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
	ios, _, errOut := newTestIO()

	err := printCheckResult(ios, "", "v0.2.0", "v0.2.0", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(errOut.String(), "Already up to date") {
		t.Errorf("expected 'Already up to date', got: %s", errOut.String())
	}
}

func TestConfirmUpdate_SkipConfirm(t *testing.T) {
	ios, _, _ := newTestIO()
	cancelled := confirmUpdate(ios, true)
	if cancelled {
		t.Error("expected not cancelled when skipConfirm=true")
	}
}

func TestConfirmUpdate_NonTTY(t *testing.T) {
	// Non-TTY IOStreams (default from newTestIO) should skip confirmation
	ios, _, _ := newTestIO()
	cancelled := confirmUpdate(ios, false)
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

func TestFallbackSource_UsesPrimary(t *testing.T) {
	errOut := &bytes.Buffer{}
	primary := &mockSource{releases: []selfupdate.SourceRelease{}}

	src := &fallbackSource{
		primary:        primary,
		fallback:       &mockSource{},
		primaryTimeout: 5 * time.Second,
		errOut:         errOut,
	}

	_, err := src.ListReleases(context.Background(), selfupdate.ParseSlug("owner/repo"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errOut.Len() > 0 {
		t.Errorf("expected no output when primary works, got: %s", errOut.String())
	}
}

func TestFallbackSource_FallsBack(t *testing.T) {
	errOut := &bytes.Buffer{}

	src := &fallbackSource{
		primary:        &mockSource{listErr: io.ErrUnexpectedEOF},
		fallback:       &mockSource{releases: []selfupdate.SourceRelease{}},
		primaryTimeout: 5 * time.Second,
		errOut:         errOut,
	}

	_, err := src.ListReleases(context.Background(), selfupdate.ParseSlug("owner/repo"))
	if err != nil {
		t.Fatalf("expected fallback to succeed, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "alternate") {
		t.Errorf("expected alternate source message, got: %s", errOut.String())
	}
}

func TestFallbackSource_BothFail(t *testing.T) {
	src := &fallbackSource{
		primary:        &mockSource{listErr: io.ErrUnexpectedEOF},
		fallback:       &mockSource{listErr: io.EOF},
		primaryTimeout: 5 * time.Second,
		errOut:         &bytes.Buffer{},
	}

	_, err := src.ListReleases(context.Background(), selfupdate.ParseSlug("owner/repo"))
	if err == nil {
		t.Fatal("expected error when both sources fail")
	}
}

func TestNewSource_ReturnsFallbackSource(t *testing.T) {
	src, err := newSource(&bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := src.(*fallbackSource); !ok {
		t.Fatal("expected fallbackSource")
	}
}
