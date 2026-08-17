package iostreams

import (
	"bytes"
	"encoding/json"
)

// Status values emitted alongside a rewritten numeric field when the raw value
// is a sentinel rather than a real measurement.
const (
	// StatusNoMeasurement means the device reported the sample but had no
	// measured value for it. The numeric field is emitted as null.
	StatusNoMeasurement = "no-measurement"
	// StatusSuspectedTimeout means the raw value matches a value believed to be
	// a firmware probe-timeout clamp. The original number is preserved.
	StatusSuspectedTimeout = "suspected-timeout"
)

// noMeasurementSentinel is the value devices report when a metric was not
// measured. Confirmed via incloud-portal, which filters it out in five places
// (apps/network .../uplink, .../overview/Uplink, .../Uplink/LinkProfile/*).
const noMeasurementSentinel = -1

// suspectedTimeoutMicros lists latency values (microseconds) believed to be a
// firmware ping-timeout clamp rather than a real round-trip measurement.
//
// Provenance: derived from Langfuse session statistics — the values are exact
// integers, land on common ping timeout steps (2s / 4s), and differ by device
// model. This has NOT been confirmed by the firmware team (IM-3180 decision
// record #4), which is why the original value is preserved and only annotated.
// Tighten (or drop) this list once firmware confirms the semantics.
//
// This is the single source of truth for the suspected-sentinel set; do not
// duplicate these numbers elsewhere.
var suspectedTimeoutMicros = []float64{2000000, 4000000}

// FieldRewrite describes how one field is renamed and annotated in structured
// (json / yaml / --jq) output.
type FieldRewrite struct {
	// To is the new field name, carrying an explicit unit suffix.
	To string
	// StatusKey is the companion field that carries the sentinel annotation.
	// Empty disables sentinel annotation entirely.
	StatusKey string
	// NoMeasurement maps the -1 sentinel to null + StatusNoMeasurement.
	NoMeasurement bool
	// SuspectedTimeout annotates values in suspectedTimeoutMicros with
	// StatusSuspectedTimeout while preserving the original number.
	SuspectedTimeout bool
}

// FieldRewrites maps the field name as returned by the platform API to its
// rewrite rule. Matching is by field name at any nesting depth.
type FieldRewrites map[string]FieldRewrite

// LatencyJitterRewrites renames the microsecond-valued latency/jitter fields so
// that a caller reading -o json can tell the unit from the output alone, and
// annotates the known sentinel values.
//
// jitter deliberately does not get the suspected-timeout treatment: there is no
// evidence that jitter is clamped on probe timeout. It does get the -1 handling,
// since a negative jitter is physically impossible and can only be the same
// "no measured value" convention.
var LatencyJitterRewrites = FieldRewrites{
	"latency": {
		To:               "latencyUs",
		StatusKey:        "latencyStatus",
		NoMeasurement:    true,
		SuspectedTimeout: true,
	},
	"jitter": {
		To:            "jitterUs",
		StatusKey:     "jitterStatus",
		NoMeasurement: true,
	},
}

// OfflineDurationRewrites renames the second-valued offline duration fields.
// No sentinel values are known for these.
var OfflineDurationRewrites = FieldRewrites{
	"totalOfflineDuration": {To: "totalOfflineDurationSeconds"},
	"avgOfflineDuration":   {To: "avgOfflineDurationSeconds"},
	"maxOfflineDuration":   {To: "maxOfflineDurationSeconds"},
}

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

	type pendingStatus struct {
		key    string
		values []interface{}
	}
	var pending []pendingStatus

	for idx, col := range cols {
		name, ok := col.(string)
		if !ok {
			continue
		}
		rule, ok := rules[name]
		if !ok {
			continue
		}
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

	for _, p := range pending {
		cols = append(cols, p.key)
		for r, row := range rows {
			cells, ok := row.([]interface{})
			if !ok {
				continue
			}
			rows[r] = append(cells, p.values[r])
		}
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
	if rule.NoMeasurement && num == noMeasurementSentinel {
		return nil, StatusNoMeasurement
	}
	if rule.SuspectedTimeout {
		for _, s := range suspectedTimeoutMicros {
			if num == s {
				return raw, StatusSuspectedTimeout
			}
		}
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
