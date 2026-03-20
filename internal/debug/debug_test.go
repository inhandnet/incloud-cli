package debug

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLog_Disabled(t *testing.T) {
	Enabled = false
	// Capture stderr
	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	Log("should not appear: %s", "test")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = origStderr

	if buf.Len() != 0 {
		t.Errorf("expected no output when disabled, got: %s", buf.String())
	}
}

func TestLog_Enabled(t *testing.T) {
	Enabled = true
	defer func() { Enabled = false }()

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w

	Log("hello %s", "world")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stderr = origStderr

	got := buf.String()
	if !strings.Contains(got, "[debug] hello world") {
		t.Errorf("expected [debug] prefix, got: %s", got)
	}
}
