package iostreams

import (
	"bytes"
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
