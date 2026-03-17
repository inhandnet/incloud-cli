package iostreams

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
)

// FormatTable renders JSON data as a table via TablePrinter.
// It extracts the "result" field if present, then:
//   - array of objects → column headers + rows
//   - single object → KEY / VALUE pairs
//
// columns filters and orders which fields to show (empty = all).
func FormatTable(data []byte, io *IOStreams, columns []string) error {
	if !gjson.ValidBytes(data) {
		// Not JSON, just print raw
		fmt.Fprintln(io.Out, string(data))
		return nil
	}

	raw := gjson.ParseBytes(data)

	// Unwrap common API envelope: {"result": ...}
	items := raw.Get("result")
	if !items.Exists() {
		items = raw
	}

	tp := NewTablePrinter(io.Out, io.IsStdoutTTY())

	switch {
	case items.IsArray():
		arr := items.Array()

		// Print pagination header for TTY table output
		if io.IsStdoutTTY() {
			if header := paginationHeader(&raw, len(arr), io); header != "" {
				fmt.Fprintln(io.Out, header)
			}
		}

		if len(arr) == 0 {
			fmt.Fprintln(io.Out, "No results.")
			return nil
		}
		renderArray(tp, arr, columns, io)
	case items.IsObject():
		renderObject(tp, &items, columns, io)
	default:
		// Scalar value
		fmt.Fprintln(io.Out, formatResult(&items))
		return nil
	}

	return tp.Render()
}

// paginationHeader builds a "Showing X of Y items (Page M of N)" line
// from the API envelope metadata. Returns "" if no pagination info is available.
func paginationHeader(raw *gjson.Result, count int, io *IOStreams) string {
	if !raw.IsObject() || count == 0 {
		return ""
	}

	c := NewColorizer(io.TermOutput())

	total := raw.Get("total")
	totalPages := raw.Get("totalPages")
	page := raw.Get("page")

	var parts []string

	if total.Exists() && total.Int() > 0 {
		parts = append(parts, fmt.Sprintf("Showing %s of %s results",
			c.Bold(fmt.Sprintf("%d", count)),
			c.Bold(fmt.Sprintf("%d", total.Int()))))
	} else {
		parts = append(parts, fmt.Sprintf("Showing %s results",
			c.Bold(fmt.Sprintf("%d", count))))
	}

	if totalPages.Exists() && totalPages.Int() > 0 {
		parts = append(parts, fmt.Sprintf("(Page %d of %d)", page.Int()+1, totalPages.Int()))
	} else if page.Exists() {
		parts = append(parts, fmt.Sprintf("(Page %d)", page.Int()+1))
	}

	return c.Gray(strings.Join(parts, " "))
}

// renderArray renders an array of gjson results as a table with header row.
func renderArray(tp *TablePrinter, items []gjson.Result, columns []string, io *IOStreams) {
	first := items[0]
	if !first.IsObject() {
		// Array of scalars
		for i := range items {
			tp.AddRow(formatResult(&items[i]))
		}
		return
	}

	cols := columns
	if len(cols) == 0 {
		cols = flattenKeys(&first)
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
		if !item.IsObject() {
			continue
		}
		row := make([]string, len(cols))
		for j, col := range cols {
			v := item.Get(col)
			row[j] = formatResult(&v)
		}
		tp.AddRow(row...)
	}
}

// renderObject renders a single object as KEY / VALUE pairs.
func renderObject(tp *TablePrinter, obj *gjson.Result, columns []string, io *IOStreams) {
	cols := columns
	if len(cols) == 0 {
		cols = flattenKeys(obj)
	}

	c := NewColorizer(io.TermOutput())
	for _, key := range cols {
		val := obj.Get(key)
		if !val.Exists() {
			continue
		}
		k := key
		if io.IsStdoutTTY() {
			k = c.Bold(key)
		}
		tp.AddRow(k, formatResult(&val))
	}
}

// formatResult converts a gjson.Result to a display string.
func formatResult(r *gjson.Result) string {
	switch r.Type {
	case gjson.Null:
		return ""
	case gjson.String:
		return r.Str
	case gjson.True:
		return "true"
	case gjson.False:
		return "false"
	case gjson.Number:
		if r.Num == float64(int64(r.Num)) {
			return fmt.Sprintf("%d", int64(r.Num))
		}
		return fmt.Sprintf("%g", r.Num)
	case gjson.JSON:
		return r.Raw
	default:
		return r.String()
	}
}

// escapeGjsonKey escapes dots and wildcards in a key segment so gjson
// treats it as a literal key rather than a nested path.
func escapeGjsonKey(s string) string {
	if !strings.ContainsAny(s, ".?*\\") {
		return s
	}
	var b strings.Builder
	for _, c := range s {
		switch c {
		case '.', '?', '*', '\\':
			b.WriteByte('\\')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// flattenKeys recursively collects leaf-level gjson-escaped dot-paths from a
// gjson object. Nested objects are expanded; arrays and scalars are leaves.
func flattenKeys(r *gjson.Result) []string {
	keys := flattenKeysWithPrefix(r, "")
	sort.Strings(keys)
	return keys
}

func flattenKeysWithPrefix(r *gjson.Result, prefix string) []string {
	var keys []string
	r.ForEach(func(key, value gjson.Result) bool {
		seg := escapeGjsonKey(key.Str)
		path := seg
		if prefix != "" {
			path = prefix + "." + seg
		}
		if value.IsObject() {
			keys = append(keys, flattenKeysWithPrefix(&value, path)...)
		} else {
			keys = append(keys, path)
		}
		return true
	})
	return keys
}
