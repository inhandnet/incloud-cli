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

// --- resolveField tests ---

func TestResolveField_TopLevel(t *testing.T) {
	obj := map[string]any{"name": "alice"}
	got := resolveField(obj, "name")
	if got != "alice" {
		t.Errorf("expected alice, got %v", got)
	}
}

func TestResolveField_DotPath(t *testing.T) {
	obj := map[string]any{
		"cpu": map[string]any{"usage": 0.5},
	}
	got := resolveField(obj, "cpu.usage")
	if got != 0.5 {
		t.Errorf("expected 0.5, got %v", got)
	}
}

func TestResolveField_DeepPath(t *testing.T) {
	obj := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
		},
	}
	got := resolveField(obj, "a.b.c")
	if got != "deep" {
		t.Errorf("expected deep, got %v", got)
	}
}

func TestResolveField_MissingKey(t *testing.T) {
	obj := map[string]any{"name": "alice"}
	got := resolveField(obj, "age")
	if got != nil {
		t.Errorf("expected nil for missing key, got %v", got)
	}
}

func TestResolveField_MissingNestedKey(t *testing.T) {
	obj := map[string]any{
		"cpu": map[string]any{"usage": 0.5},
	}
	got := resolveField(obj, "cpu.temp")
	if got != nil {
		t.Errorf("expected nil for missing nested key, got %v", got)
	}
}

func TestResolveField_PathThroughScalar(t *testing.T) {
	obj := map[string]any{"name": "alice"}
	got := resolveField(obj, "name.sub")
	if got != nil {
		t.Errorf("expected nil when path goes through scalar, got %v", got)
	}
}

func TestResolveField_PathThroughArray(t *testing.T) {
	obj := map[string]any{
		"tags": []any{"a", "b"},
	}
	got := resolveField(obj, "tags.0")
	if got != nil {
		t.Errorf("expected nil when path goes through array, got %v", got)
	}
}

func TestResolveField_ReturnsNestedObject(t *testing.T) {
	obj := map[string]any{
		"cpu": map[string]any{"usage": 0.5, "cores": float64(4)},
	}
	got := resolveField(obj, "cpu")
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", got)
	}
	if m["usage"] != 0.5 {
		t.Errorf("expected usage=0.5, got %v", m["usage"])
	}
}

func TestResolveField_ReturnsArray(t *testing.T) {
	obj := map[string]any{
		"tags": []any{"a", "b"},
	}
	got := resolveField(obj, "tags")
	arr, ok := got.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", got)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr))
	}
}

// --- formatValue tests ---

func TestFormatValue_Nil(t *testing.T) {
	if got := formatValue(nil); got != "" {
		t.Errorf("expected empty string for nil, got %q", got)
	}
}

func TestFormatValue_String(t *testing.T) {
	if got := formatValue("hello"); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestFormatValue_EmptyString(t *testing.T) {
	if got := formatValue(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFormatValue_Integer(t *testing.T) {
	// JSON numbers are float64
	if got := formatValue(float64(42)); got != "42" {
		t.Errorf("expected 42, got %q", got)
	}
}

func TestFormatValue_NegativeInteger(t *testing.T) {
	if got := formatValue(float64(-7)); got != "-7" {
		t.Errorf("expected -7, got %q", got)
	}
}

func TestFormatValue_Zero(t *testing.T) {
	if got := formatValue(float64(0)); got != "0" {
		t.Errorf("expected 0, got %q", got)
	}
}

func TestFormatValue_Float(t *testing.T) {
	if got := formatValue(3.14); got != "3.14" {
		t.Errorf("expected 3.14, got %q", got)
	}
}

func TestFormatValue_SmallFloat(t *testing.T) {
	if got := formatValue(0.024570024); got != "0.024570024" {
		t.Errorf("expected 0.024570024, got %q", got)
	}
}

func TestFormatValue_BoolTrue(t *testing.T) {
	if got := formatValue(true); got != "true" {
		t.Errorf("expected true, got %q", got)
	}
}

func TestFormatValue_BoolFalse(t *testing.T) {
	if got := formatValue(false); got != "false" {
		t.Errorf("expected false, got %q", got)
	}
}

func TestFormatValue_NestedObject(t *testing.T) {
	obj := map[string]any{"usage": 0.5}
	got := formatValue(obj)
	if got != `{"usage":0.5}` {
		t.Errorf("expected compact JSON, got %q", got)
	}
}

func TestFormatValue_Array(t *testing.T) {
	arr := []any{"a", "b"}
	got := formatValue(arr)
	if got != `["a","b"]` {
		t.Errorf("expected compact JSON array, got %q", got)
	}
}

func TestFormatValue_EmptyArray(t *testing.T) {
	arr := []any{}
	got := formatValue(arr)
	if got != `[]` {
		t.Errorf("expected [], got %q", got)
	}
}

func TestFormatValue_LargeInteger(t *testing.T) {
	if got := formatValue(float64(21114126336)); got != "21114126336" {
		t.Errorf("expected 21114126336, got %q", got)
	}
}

// --- flattenKeys tests ---

func TestFlattenKeys_Flat(t *testing.T) {
	m := map[string]any{"name": "alice", "age": float64(30)}
	got := flattenKeys(m, "")
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
	m := map[string]any{
		"cpu":    map[string]any{"usage": 0.5},
		"memory": map[string]any{"free": float64(1024), "total": float64(4096)},
	}
	got := flattenKeys(m, "")
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
	m := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "val",
			},
		},
	}
	got := flattenKeys(m, "")
	if len(got) != 1 || got[0] != "a.b.c" {
		t.Errorf("expected [a.b.c], got %v", got)
	}
}

func TestFlattenKeys_ArrayStopsFlattening(t *testing.T) {
	m := map[string]any{
		"name": "alice",
		"tags": []any{"a", "b"},
	}
	got := flattenKeys(m, "")
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
	m := map[string]any{
		"id":     "123",
		"cpu":    map[string]any{"usage": 0.5},
		"tags":   []any{"x"},
		"active": true,
		"meta":   map[string]any{"region": "us", "nested": map[string]any{"deep": float64(1)}},
	}
	got := flattenKeys(m, "")
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
