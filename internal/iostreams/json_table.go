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

	// Print pagination header for TTY table output
	if io.IsStdoutTTY() {
		if header := paginationHeader(raw, items, io); header != "" {
			fmt.Fprintln(io.Out, header)
		}
	}

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

// paginationHeader builds a "Showing X of Y items (Page M of N)" line
// from the API envelope metadata. Returns "" if no pagination info is available.
func paginationHeader(raw, items any, io *IOStreams) string {
	obj, ok := raw.(map[string]any)
	if !ok {
		return ""
	}

	// Count items on current page
	arr, ok := items.([]any)
	if !ok {
		return ""
	}
	count := len(arr)
	if count == 0 {
		return ""
	}

	c := NewColorizer(io.TermOutput())

	total := intFromJSON(obj, "total")
	totalPages := intFromJSON(obj, "totalPages")
	page := intFromJSON(obj, "page")

	var parts []string

	if total > 0 {
		parts = append(parts, fmt.Sprintf("Showing %s of %s results",
			c.Bold(fmt.Sprintf("%d", count)),
			c.Bold(fmt.Sprintf("%d", total))))
	} else {
		parts = append(parts, fmt.Sprintf("Showing %s results",
			c.Bold(fmt.Sprintf("%d", count))))
	}

	if totalPages > 0 {
		parts = append(parts, fmt.Sprintf("(Page %d of %d)", page+1, totalPages))
	} else if page >= 0 {
		parts = append(parts, fmt.Sprintf("(Page %d)", page+1))
	}

	return c.Gray(strings.Join(parts, " "))
}

// intFromJSON extracts an integer value from a JSON object by key.
// Returns -1 if the key is missing or not a number.
func intFromJSON(obj map[string]any, key string) int {
	v, ok := obj[key]
	if !ok {
		return -1
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return -1
	}
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
			row[j] = formatValue(resolveField(obj, col))
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
		val := resolveField(obj, key)
		if val == nil {
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

// resolveField walks a dot-separated path (e.g. "actor.name") into nested maps.
func resolveField(obj map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = obj
	for _, p := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[p]
	}
	return current
}

// flattenKeys recursively collects leaf-level dot-paths from a nested map.
// Nested maps are expanded; arrays and scalars are treated as leaves.
func flattenKeys(m map[string]any, prefix string) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		if sub, ok := v.(map[string]any); ok {
			keys = append(keys, flattenKeys(sub, path)...)
		} else {
			keys = append(keys, path)
		}
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(m map[string]any) []string {
	return flattenKeys(m, "")
}
