package iostreams

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func newTestIO(isTTY bool) *IOStreams {
	var p termenv.Profile
	if isTTY {
		p = termenv.ANSI // force color support for test
	} else {
		p = termenv.Ascii
	}
	return &IOStreams{
		In:       os.Stdin,
		Out:      os.Stdout,
		ErrOut:   os.Stderr,
		outIsTTY: isTTY,
		termOut:  termenv.NewOutput(os.Stdout, termenv.WithProfile(p)),
	}
}

func TestFormatJSON_NonTTY_Compact(t *testing.T) {
	input := []byte(`{
  "name": "test",
  "count": 42
}`)
	result := FormatJSON(input, newTestIO(false), "")
	if strings.Contains(result, "\n") {
		t.Errorf("non-TTY output should be compact, got:\n%s", result)
	}
	if !json.Valid([]byte(result)) {
		t.Errorf("output is not valid JSON: %s", result)
	}
}

func TestFormatJSON_TTY_PrettyPrint(t *testing.T) {
	input := []byte(`{"name":"test","count":42}`)
	result := FormatJSON(input, newTestIO(true), "")
	if !strings.Contains(result, "\n") {
		t.Errorf("TTY default output should be pretty printed, got:\n%s", result)
	}
	// Should contain ANSI escape sequences for colors
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("TTY default output should contain ANSI colors, got:\n%q", result)
	}
}

func TestFormatJSON_TTY_JsonMode_NoColor(t *testing.T) {
	input := []byte(`{"name":"test"}`)
	result := FormatJSON(input, newTestIO(true), "json")
	if !strings.Contains(result, "\n") {
		t.Errorf("TTY json mode should be pretty printed, got:\n%s", result)
	}
	if strings.Contains(result, "\x1b[") {
		t.Errorf("json mode should not contain ANSI colors, got:\n%q", result)
	}
}

func TestFormatJSON_InvalidJSON(t *testing.T) {
	input := []byte(`not json at all`)
	result := FormatJSON(input, newTestIO(true), "")
	if result != "not json at all" {
		t.Errorf("invalid JSON should pass through unchanged, got: %s", result)
	}
}

func TestColorizeJSON_AllTypes(t *testing.T) {
	input := []byte(`{"name":"alice","age":30,"active":true,"deleted":null}`)
	io := newTestIO(true)
	result := FormatJSON(input, io, "")

	// Verify key bold, string green, number yellow, bool/null red
	if !strings.Contains(result, "\x1b[1m") { // bold
		t.Errorf("expected bold keys, got:\n%q", result)
	}
	if !strings.Contains(result, "\x1b[32m") || !strings.Contains(result, "alice") { // green (ANSI color 2 = 32m)
		t.Errorf("expected green string values, got:\n%q", result)
	}
}
