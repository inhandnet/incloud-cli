package iostreams

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/muesli/termenv"
	"github.com/tidwall/gjson"
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

// --- formatResult tests (gjson-based) ---

func TestFormatResult_Null(t *testing.T) {
	r := gjson.Parse("null")
	if got := formatResult(&r); got != "" {
		t.Errorf("expected empty string for null, got %q", got)
	}
}

func TestFormatResult_String(t *testing.T) {
	r := gjson.Parse(`"hello"`)
	if got := formatResult(&r); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestFormatResult_EmptyString(t *testing.T) {
	r := gjson.Parse(`""`)
	if got := formatResult(&r); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormatResult_Integer(t *testing.T) {
	r := gjson.Parse("42")
	if got := formatResult(&r); got != "42" {
		t.Errorf("expected 42, got %q", got)
	}
}

func TestFormatResult_NegativeInteger(t *testing.T) {
	r := gjson.Parse("-7")
	if got := formatResult(&r); got != "-7" {
		t.Errorf("expected -7, got %q", got)
	}
}

func TestFormatResult_Zero(t *testing.T) {
	r := gjson.Parse("0")
	if got := formatResult(&r); got != "0" {
		t.Errorf("expected 0, got %q", got)
	}
}

func TestFormatResult_Float(t *testing.T) {
	r := gjson.Parse("3.14")
	if got := formatResult(&r); got != "3.14" {
		t.Errorf("expected 3.14, got %q", got)
	}
}

func TestFormatResult_SmallFloat(t *testing.T) {
	r := gjson.Parse("0.024570024")
	if got := formatResult(&r); got != "0.025" {
		t.Errorf("expected 0.025, got %q", got)
	}
}

func TestFormatResult_BoolTrue(t *testing.T) {
	r := gjson.Parse("true")
	if got := formatResult(&r); got != "true" {
		t.Errorf("expected true, got %q", got)
	}
}

func TestFormatResult_BoolFalse(t *testing.T) {
	r := gjson.Parse("false")
	if got := formatResult(&r); got != "false" {
		t.Errorf("expected false, got %q", got)
	}
}

func TestFormatResult_NestedObject(t *testing.T) {
	r := gjson.Parse(`{"usage":0.5}`)
	got := formatResult(&r)
	if got != `{"usage":0.5}` {
		t.Errorf("expected compact JSON, got %q", got)
	}
}

func TestFormatResult_Array(t *testing.T) {
	r := gjson.Parse(`["a","b"]`)
	got := formatResult(&r)
	if got != `["a","b"]` {
		t.Errorf("expected compact JSON array, got %q", got)
	}
}

func TestFormatResult_EmptyArray(t *testing.T) {
	r := gjson.Parse(`[]`)
	got := formatResult(&r)
	if got != `[]` {
		t.Errorf("expected [], got %q", got)
	}
}

func TestFormatResult_LargeInteger(t *testing.T) {
	r := gjson.Parse("21114126336")
	if got := formatResult(&r); got != "21114126336" {
		t.Errorf("expected 21114126336, got %q", got)
	}
}

// --- gjson dot-path resolution tests ---

func TestGjsonGet_TopLevel(t *testing.T) {
	r := gjson.Parse(`{"name":"alice"}`)
	got := r.Get("name").Str
	if got != "alice" {
		t.Errorf("expected alice, got %v", got)
	}
}

func TestGjsonGet_DotPath(t *testing.T) {
	r := gjson.Parse(`{"cpu":{"usage":0.5}}`)
	got := r.Get("cpu.usage").Num
	if got != 0.5 {
		t.Errorf("expected 0.5, got %v", got)
	}
}

func TestGjsonGet_DeepPath(t *testing.T) {
	r := gjson.Parse(`{"a":{"b":{"c":"deep"}}}`)
	got := r.Get("a.b.c").Str
	if got != "deep" {
		t.Errorf("expected deep, got %v", got)
	}
}

func TestGjsonGet_MissingKey(t *testing.T) {
	r := gjson.Parse(`{"name":"alice"}`)
	got := r.Get("age")
	if got.Exists() {
		t.Errorf("expected non-existent for missing key, got %v", got)
	}
}

func TestGjsonGet_MissingNestedKey(t *testing.T) {
	r := gjson.Parse(`{"cpu":{"usage":0.5}}`)
	got := r.Get("cpu.temp")
	if got.Exists() {
		t.Errorf("expected non-existent for missing nested key, got %v", got)
	}
}

func TestGjsonGet_ReturnsNestedObject(t *testing.T) {
	r := gjson.Parse(`{"cpu":{"usage":0.5,"cores":4}}`)
	got := r.Get("cpu")
	if !got.IsObject() {
		t.Fatalf("expected object, got %v", got.Type)
	}
	if got.Get("usage").Num != 0.5 {
		t.Errorf("expected usage=0.5, got %v", got.Get("usage").Num)
	}
}

func TestGjsonGet_ReturnsArray(t *testing.T) {
	r := gjson.Parse(`{"tags":["a","b"]}`)
	got := r.Get("tags")
	if !got.IsArray() {
		t.Fatalf("expected array, got %v", got.Type)
	}
	if len(got.Array()) != 2 {
		t.Errorf("expected 2 elements, got %d", len(got.Array()))
	}
}

// --- flattenKeys tests (gjson-based) ---

func TestFlattenKeys_Flat(t *testing.T) {
	r := gjson.Parse(`{"age":30,"name":"alice"}`)
	got := flattenKeys(&r)
	expect := []string{"age", "name"}
	if len(got) != len(expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: expected %s, got %s", i, expect[i], got[i])
		}
	}
}

func TestFlattenKeys_Nested(t *testing.T) {
	r := gjson.Parse(`{"cpu":{"usage":0.5},"memory":{"free":1024,"total":4096}}`)
	got := flattenKeys(&r)
	expect := []string{"cpu.usage", "memory.free", "memory.total"}
	if len(got) != len(expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: expected %s, got %s", i, expect[i], got[i])
		}
	}
}

func TestFlattenKeys_DeepNested(t *testing.T) {
	r := gjson.Parse(`{"a":{"b":{"c":"val"}}}`)
	got := flattenKeys(&r)
	if len(got) != 1 || got[0] != "a.b.c" {
		t.Errorf("expected [a.b.c], got %v", got)
	}
}

func TestFlattenKeys_ArrayStopsFlattening(t *testing.T) {
	r := gjson.Parse(`{"name":"alice","tags":["a","b"]}`)
	got := flattenKeys(&r)
	expect := []string{"name", "tags"}
	if len(got) != len(expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: expected %s, got %s", i, expect[i], got[i])
		}
	}
}

func TestFlattenKeys_Mixed(t *testing.T) {
	r := gjson.Parse(`{"active":true,"cpu":{"usage":0.5},"id":"123","meta":{"nested":{"deep":1},"region":"us"},"tags":["x"]}`)
	got := flattenKeys(&r)
	expect := []string{"active", "cpu.usage", "id", "meta.nested.deep", "meta.region", "tags"}
	if len(got) != len(expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: expected %s, got %s", i, expect[i], got[i])
		}
	}
}

func TestFlattenKeys_DotInKey(t *testing.T) {
	// Keys containing literal dots must be escaped so gjson.Get resolves them correctly.
	r := gjson.Parse(`{"a.b":"dotval","normal":"ok"}`)
	got := flattenKeys(&r)
	expect := []string{`a\.b`, "normal"}
	if len(got) != len(expect) {
		t.Fatalf("expected %v, got %v", expect, got)
	}
	for i := range expect {
		if got[i] != expect[i] {
			t.Errorf("index %d: expected %s, got %s", i, expect[i], got[i])
		}
	}
	// Verify the escaped path round-trips through gjson.Get
	val := r.Get(got[0])
	if val.Str != "dotval" {
		t.Errorf("expected gjson.Get(%q) = dotval, got %q", got[0], val.Str)
	}
}

func TestFormatTable_DotInKey(t *testing.T) {
	// End-to-end: a key with a literal dot should render correctly in table output.
	data := []byte(`{"result":[{"a.b":"val1","c":"val2"},{"a.b":"val3","c":"val4"}]}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "val1") || !strings.Contains(out, "val3") {
		t.Errorf("expected dot-key values in output, got:\n%s", out)
	}
}

// --- FormatTable integration: auto-flatten object without columns ---

func TestFormatTable_Object_AutoFlatten(t *testing.T) {
	data := []byte(`{"result":{"cpu":{"usage":0.5},"memory":{"free":1024,"total":4096},"updatedAt":"2026-01-01"}}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Should show flattened dot-path keys, not JSON blobs
	lines := strings.Split(strings.TrimSpace(out), "\n")
	expected := map[string]string{
		"cpu.usage":    "0.5",
		"memory.free":  "1024",
		"memory.total": "4096",
		"updatedAt":    "2026-01-01",
	}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d:\n%s", len(expected), len(lines), out)
	}
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			t.Errorf("expected TSV line, got: %q", line)
			continue
		}
		key, val := parts[0], parts[1]
		if ev, ok := expected[key]; ok {
			if val != ev {
				t.Errorf("key %s: expected %s, got %s", key, ev, val)
			}
		} else {
			t.Errorf("unexpected key: %s", key)
		}
	}
}

func TestFormatTable_Object_AutoFlatten_WithArray(t *testing.T) {
	data := []byte(`{"result":{"name":"dev1","tags":["a","b"],"cpu":{"usage":0.5}}}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// tags should stay as JSON array, cpu should flatten
	if !strings.Contains(out, "cpu.usage\t0.5") {
		t.Errorf("expected cpu.usage flattened, got:\n%s", out)
	}
	if !strings.Contains(out, "name\tdev1") {
		t.Errorf("expected name scalar, got:\n%s", out)
	}
	if !strings.Contains(out, `tags	["a","b"]`) {
		t.Errorf("expected tags as JSON array, got:\n%s", out)
	}
}

// --- FormatTable integration: non-JSON input ---

func TestFormatTable_NonJSON(t *testing.T) {
	data := []byte("plain text output")
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "plain text output") {
		t.Errorf("expected raw passthrough, got:\n%s", out)
	}
}

// --- FormatTable integration: scalar result ---

func TestFormatTable_ScalarResult(t *testing.T) {
	data := []byte(`{"result":"hello"}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "hello" {
		t.Errorf("expected hello, got %q", out)
	}
}

func TestFormatTable_NumericResult(t *testing.T) {
	data := []byte(`{"result":42}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "42" {
		t.Errorf("expected 42, got %q", out)
	}
}

// --- FormatTable integration: object with columns filter ---

func TestFormatTable_Object_WithColumns(t *testing.T) {
	data := []byte(`{"result":{"name":"alice","age":30,"role":"admin"}}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, []string{"name", "role"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice") || !strings.Contains(out, "admin") {
		t.Errorf("expected name and role, got:\n%s", out)
	}
	if strings.Contains(out, "30") {
		t.Errorf("age should be filtered out, got:\n%s", out)
	}
}

// --- FormatTable integration: object with dot-path columns ---

func TestFormatTable_Object_DotPathColumns(t *testing.T) {
	data := []byte(`{"result":{"cpu":{"usage":0.5},"memory":{"free":1024,"total":4096}}}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, []string{"cpu.usage", "memory.free", "memory.total"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "0.5") {
		t.Errorf("expected cpu.usage=0.5, got:\n%s", out)
	}
	if !strings.Contains(out, "1024") || !strings.Contains(out, "4096") {
		t.Errorf("expected memory values, got:\n%s", out)
	}
}

// --- FormatTable integration: array of scalars ---

func TestFormatTable_ArrayOfScalars(t *testing.T) {
	data := []byte(`{"result":["one","two","three"]}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "one") || !strings.Contains(out, "two") || !strings.Contains(out, "three") {
		t.Errorf("expected scalar values, got:\n%s", out)
	}
}

// --- FormatTable integration: no envelope (raw object) ---

func TestFormatTable_NoEnvelope(t *testing.T) {
	data := []byte(`{"name":"alice","age":30}`)
	io, buf := newTestIOWithBuf(false)
	if err := FormatTable(data, io, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice") || !strings.Contains(out, "30") {
		t.Errorf("expected key-value pairs, got:\n%s", out)
	}
}
