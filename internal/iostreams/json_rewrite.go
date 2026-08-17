package iostreams

import (
	"bytes"
	"encoding/json"
)

// StatusTimeout is emitted alongside a rewritten numeric field when the raw
// value is the timeout sentinel rather than a measurement. The device firmware
// reports -1 when a probe timed out (nezha-agent/pkg/message/message.go:117:
// "latency, us, set value to -1 if timeout"); the numeric field is emitted as
// null so the sentinel never reaches a consumer as a number.
//
// This is the only sentinel on this link. In particular there is no clamp
// ceiling: values such as 2000000 (2s) are real measurements on a degraded
// cellular link and must never be annotated.
const StatusTimeout = "timeout"

// timeoutSentinel is the value devices report when a probe timed out.
const timeoutSentinel = -1

// FieldRewrite describes how one field is renamed and annotated in structured
// (json / yaml / --jq) output.
type FieldRewrite struct {
	// To is the new field name, carrying an explicit unit suffix.
	To string
	// StatusKey is the companion field that carries the sentinel annotation.
	// Empty disables sentinel annotation entirely.
	StatusKey string
	// Timeout maps the -1 sentinel to null + StatusTimeout.
	Timeout bool
}

// FieldRewrites maps the field name as returned by the platform API to its
// rewrite rule. Matching is by field name at any nesting depth, plus by column
// name for the columnar time-series shape.
type FieldRewrites map[string]FieldRewrite

// applyFieldRewrites renames and annotates fields anywhere in the JSON body.
// Invalid JSON is returned unchanged.
func applyFieldRewrites(data []byte, rules FieldRewrites) []byte {
	if len(rules) == 0 {
		return data
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return data
	}
	v = rewriteValue(v, rules)
	out, err := json.Marshal(v)
	if err != nil {
		return data
	}
	return out
}

// rewriteValue walks the decoded JSON tree, applying rules to every matching
// object key and to columnar time-series objects.
func rewriteValue(v interface{}, rules FieldRewrites) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, child := range t {
			t[k] = rewriteValue(child, rules)
		}
		rewriteSeriesObject(t, rules)
		rewriteObjectKeys(t, rules)
		return t
	case []interface{}:
		for i, child := range t {
			t[i] = rewriteValue(child, rules)
		}
		return t
	default:
		return v
	}
}

// rewriteObjectKeys renames matching keys of a single object and attaches the
// status companion field where the raw value is a sentinel.
func rewriteObjectKeys(obj map[string]interface{}, rules FieldRewrites) {
	for from, rule := range rules {
		raw, ok := obj[from]
		if !ok {
			continue
		}
		delete(obj, from)
		value, status := applyRule(raw, rule)
		obj[rule.To] = value
		if status != "" && rule.StatusKey != "" {
			obj[rule.StatusKey] = status
		}
	}
}

// rewriteSeriesObject handles the columnar time-series shape returned by the
// perf-trend style endpoints, where field names live in a "columns" (or
// "fields") array and the numbers live positionally in "values" (or "data").
// A status column is appended only when at least one row carries a sentinel.
//
// Every row that is an array is normalized to the width of the column list, so
// a short input row cannot shift the appended status values into the wrong
// column.
func rewriteSeriesObject(obj map[string]interface{}, rules FieldRewrites) {
	colKey, rowKey := seriesKeys(obj)
	if colKey == "" {
		return
	}
	cols, _ := obj[colKey].([]interface{})
	rows, _ := obj[rowKey].([]interface{})
	if len(cols) == 0 {
		return
	}
	width := len(cols)

	type pendingStatus struct {
		key    string
		values []interface{}
	}
	var pending []pendingStatus
	matched := false

	for idx, col := range cols {
		name, ok := col.(string)
		if !ok {
			continue
		}
		rule, ok := rules[name]
		if !ok {
			continue
		}
		matched = true
		cols[idx] = rule.To

		statuses := make([]interface{}, len(rows))
		anyStatus := false
		for r, row := range rows {
			cells, ok := row.([]interface{})
			if !ok || idx >= len(cells) {
				continue
			}
			value, status := applyRule(cells[idx], rule)
			cells[idx] = value
			if status != "" {
				statuses[r] = status
				anyStatus = true
			}
		}
		if anyStatus && rule.StatusKey != "" {
			pending = append(pending, pendingStatus{key: rule.StatusKey, values: statuses})
		}
	}

	if !matched {
		return
	}

	for r, row := range rows {
		cells, ok := row.([]interface{})
		if !ok {
			continue
		}
		// Pad short rows first: appending a status onto a short row would
		// otherwise land it in a value column.
		for len(cells) < width {
			cells = append(cells, nil)
		}
		for _, p := range pending {
			cells = append(cells, p.values[r])
		}
		rows[r] = cells
	}
	for _, p := range pending {
		cols = append(cols, p.key)
	}

	obj[colKey] = cols
	obj[rowKey] = rows
}

// seriesKeys detects which columnar naming convention an object uses.
// Both conventions appear across backend services (see FlattenSeries).
func seriesKeys(obj map[string]interface{}) (string, string) {
	for _, pair := range [][2]string{{"columns", "values"}, {"fields", "data"}} {
		if _, ok := obj[pair[0]].([]interface{}); !ok {
			continue
		}
		if _, ok := obj[pair[1]].([]interface{}); !ok {
			continue
		}
		return pair[0], pair[1]
	}
	return "", ""
}

// applyRule maps a raw value to its rewritten value plus an optional status.
// Non-numeric values (null, strings, objects) pass through untouched.
func applyRule(raw interface{}, rule FieldRewrite) (interface{}, string) {
	num, ok := asFloat(raw)
	if !ok {
		return raw, ""
	}
	if rule.Timeout && num == timeoutSentinel {
		return nil, StatusTimeout
	}
	return raw, ""
}

// asFloat extracts a numeric value from a decoded JSON value.
func asFloat(raw interface{}) (float64, bool) {
	switch n := raw.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		return n, true
	default:
		return 0, false
	}
}
