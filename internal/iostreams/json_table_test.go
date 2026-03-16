package iostreams

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

func newTestIOWithBuf(isTTY bool) (*IOStreams, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	var p termenv.Profile
	if isTTY {
		p = termenv.ANSI
	} else {
		p = termenv.Ascii
	}
	io := &IOStreams{
		In:       os.Stdin,
		Out:      buf,
		ErrOut:   os.Stderr,
		outIsTTY: isTTY,
		termOut:  termenv.NewOutput(os.Stdout, termenv.WithProfile(p)),
	}
	return io, buf
}

func TestFormatTable_Array_TTY(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","status":"online"},{"name":"dev2","status":"offline"}]}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Should have pagination header + column header + 2 data rows
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (pagination + header + 2 rows), got %d:\n%s", len(lines), out)
	}
	// Should contain pagination info
	if !strings.Contains(out, "Showing") || !strings.Contains(out, "2") {
		t.Errorf("expected pagination header with count, got:\n%s", out)
	}
	// Header should contain column names
	if !strings.Contains(strings.ToUpper(out), "NAME") {
		t.Errorf("expected NAME header, got:\n%s", out)
	}
	if !strings.Contains(out, "dev1") || !strings.Contains(out, "dev2") {
		t.Errorf("expected device names in output, got:\n%s", out)
	}
}

func TestFormatTable_Array_NonTTY_TSV(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","status":"online"},{"name":"dev2","status":"offline"}]}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Non-TTY should be TSV (tab-separated)
	for _, line := range lines {
		if !strings.Contains(line, "\t") {
			t.Errorf("non-TTY output should be TSV, got line: %q", line)
		}
	}
}

func TestFormatTable_Object_KeyValue(t *testing.T) {
	data := []byte(`{"result":{"name":"alice","age":30}}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice") || !strings.Contains(out, "30") {
		t.Errorf("expected key-value output, got:\n%s", out)
	}
}

func TestFormatTable_Columns_Filter(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","status":"online","id":"123"},{"name":"dev2","status":"offline","id":"456"}]}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, []string{"name", "status"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Should NOT contain id column
	if strings.Contains(out, "123") || strings.Contains(out, "456") {
		t.Errorf("columns filter should exclude 'id', got:\n%s", out)
	}
	if !strings.Contains(out, "dev1") {
		t.Errorf("columns filter should include 'name', got:\n%s", out)
	}
}

func TestFormatTable_EmptyArray(t *testing.T) {
	data := []byte(`{"result":[]}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "No results") {
		t.Errorf("expected 'No results' for empty array, got:\n%s", out)
	}
}

func TestFormatTable_NestedValue(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","tags":["a","b"]}]}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Nested array should be rendered as JSON string
	if !strings.Contains(out, `["a","b"]`) {
		t.Errorf("nested value should be compact JSON, got:\n%s", out)
	}
}

func TestFormatTable_PaginationHeader_WithTotal(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1"},{"name":"dev2"}],"total":50,"totalPages":5,"page":0}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "2") || !strings.Contains(out, "50") {
		t.Errorf("expected 'Showing 2 of 50', got:\n%s", out)
	}
	if !strings.Contains(out, "Page 1 of 5") {
		t.Errorf("expected 'Page 1 of 5', got:\n%s", out)
	}
}

func TestFormatTable_PaginationHeader_WithoutTotal(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1"}]}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Showing") || !strings.Contains(out, "1") {
		t.Errorf("expected 'Showing 1 results', got:\n%s", out)
	}
}

func TestFormatTable_PaginationHeader_NonTTY_NoHeader(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1"}],"total":50,"totalPages":5,"page":0}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "Showing") {
		t.Errorf("non-TTY should not have pagination header, got:\n%s", out)
	}
}

func TestFormatTable_PaginationHeader_Page2(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1"}],"total":25,"totalPages":3,"page":1}`)
	io, buf := newTestIOWithBuf(true)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Page 2 of 3") {
		t.Errorf("expected 'Page 2 of 3' (0-indexed page=1), got:\n%s", out)
	}
}
