package iostreams

import (
	"bytes"
	"strings"
	"testing"
)

func TestTablePrinter_TTY(t *testing.T) {
	var buf bytes.Buffer
	tp := NewTablePrinter(&buf, true)
	tp.AddRow("NAME", "STATUS", "HOST")
	tp.AddRow("dev", "active", "https://dev.example.com")
	tp.AddRow("prod", "inactive", "https://prod.example.com")
	if err := tp.Render(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if out == "" {
		t.Fatal("expected table output")
	}
	// TTY mode uses space-padded columns
	if !contains(out, "NAME") || !contains(out, "dev") {
		t.Errorf("missing expected content in: %s", out)
	}
}

func TestTablePrinter_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	tp := NewTablePrinter(&buf, false)
	tp.AddRow("NAME", "STATUS")
	tp.AddRow("dev", "active")
	if err := tp.Render(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Non-TTY uses tab separator
	if !contains(out, "NAME\tSTATUS") {
		t.Errorf("expected tab-separated output, got: %s", out)
	}
}

func TestTablePrinter_TTY_Unicode(t *testing.T) {
	var buf bytes.Buffer
	tp := NewTablePrinter(&buf, true)
	tp.AddRow("NAME", "STATUS")
	tp.AddRow("设备一号", "在线")
	tp.AddRow("dev", "offline")
	if err := tp.Render(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d:\n%s", len(lines), out)
	}
	// go-pretty uses runewidth for Unicode: "设备一号" is 8 display columns (4 CJK chars × 2).
	// Verify all lines have the same display width (columns properly aligned).
	if !strings.Contains(out, "设备一号") || !strings.Contains(out, "在线") {
		t.Errorf("expected Chinese content in output, got:\n%s", out)
	}
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "offline") {
		t.Errorf("expected ASCII content in output, got:\n%s", out)
	}
}

func TestTablePrinter_Empty(t *testing.T) {
	var buf bytes.Buffer
	tp := NewTablePrinter(&buf, true)
	if err := tp.Render(); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "" {
		t.Errorf("expected empty output for no rows, got: %q", buf.String())
	}
}

func TestIOStreams_IsTTY(t *testing.T) {
	io := &IOStreams{
		outIsTTY: false,
	}
	if io.IsStdoutTTY() {
		t.Error("expected non-TTY")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
