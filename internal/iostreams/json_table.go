package iostreams

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// FormatTable renders JSON data as a table via TablePrinter.
// It extracts the "result" field if present, then:
//   - array of objects → column headers + rows
//   - single object → KEY / VALUE pairs
//
// columns filters and orders which fields to show (empty = all).
func FormatTable(data []byte, io *IOStreams, columns []string) error {
	// Parse JSON
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		// Not JSON, just print raw
		fmt.Fprintln(io.Out, string(data))
		return nil
	}

	// Unwrap common API envelope: {"result": ...}
	items := unwrapResult(raw)

	tp := NewTablePrinter(io.Out, io.IsStdoutTTY())

	switch v := items.(type) {
	case []any:
		if len(v) == 0 {
			fmt.Fprintln(io.Out, "No results.")
			return nil
		}
		renderArray(tp, v, columns, io)
	case map[string]any:
		renderObject(tp, v, columns, io)
	default:
		// Scalar value
		fmt.Fprintln(io.Out, formatValue(items))
		return nil
	}

	return tp.Render()
}

// unwrapResult extracts the "result" field from a top-level object if present.
func unwrapResult(v any) any {
	obj, ok := v.(map[string]any)
	if !ok {
		return v
	}
	if result, ok := obj["result"]; ok {
		return result
	}
	return v
}

// renderArray renders an array of objects as a table with header row.
func renderArray(tp *TablePrinter, items []any, columns []string, io *IOStreams) {
	// Collect all keys from first object to determine columns
	firstObj, ok := items[0].(map[string]any)
	if !ok {
		// Array of scalars
		for _, item := range items {
			tp.AddRow(formatValue(item))
		}
		return
	}

	cols := columns
	if len(cols) == 0 {
		cols = sortedKeys(firstObj)
	}

	// Header row (bold in TTY)
	header := make([]string, len(cols))
	c := NewColorizer(io.TermOutput())
	for i, col := range cols {
		if io.IsStdoutTTY() {
			header[i] = c.Bold(strings.ToUpper(col))
		} else {
			header[i] = strings.ToUpper(col)
		}
	}
	tp.AddRow(header...)

	// Data rows
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		row := make([]string, len(cols))
		for j, col := range cols {
			row[j] = formatValue(obj[col])
		}
		tp.AddRow(row...)
	}
}

// renderObject renders a single object as KEY / VALUE pairs.
func renderObject(tp *TablePrinter, obj map[string]any, columns []string, io *IOStreams) {
	cols := columns
	if len(cols) == 0 {
		cols = sortedKeys(obj)
	}

	c := NewColorizer(io.TermOutput())
	for _, key := range cols {
		val, ok := obj[key]
		if !ok {
			continue
		}
		k := key
		if io.IsStdoutTTY() {
			k = c.Bold(key)
		}
		tp.AddRow(k, formatValue(val))
	}
}

// formatValue converts any JSON value to a display string.
func formatValue(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		// Nested object/array: compact JSON
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
