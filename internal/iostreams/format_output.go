package iostreams

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// TransformFunc converts raw API response bytes into a format suitable for FormatTable.
type TransformFunc func([]byte) ([]byte, error)

// FormatOption configures FormatOutput behavior.
type FormatOption func(*formatOptions)

type formatOptions struct {
	transform           TransformFunc
	structuredTransform TransformFunc
	formatters          ColumnFormatters
	columns             []string
	rewrites            FieldRewrites
}

// WithTransform sets a transform function applied before table rendering.
// The transform is only used for table output; json/yaml always use the original body.
func WithTransform(fn TransformFunc) FormatOption {
	return func(o *formatOptions) {
		o.transform = fn
	}
}

// WithStructuredTransform sets a transform applied to the yaml / json / --jq
// payload, after envelope unwrapping. It is the counterpart to WithTransform:
// a command whose rendering does more than reformat — reordering samples, for
// one — has to carry that through to machine callers too, or the flag that asked
// for it silently applies to table output only.
//
// The transform sees the payload structured callers actually parse (envelope
// unwrapped, pagination normalized), not the flattened rows a table shows, so it
// must preserve that shape rather than reshape it into table rows.
func WithStructuredTransform(fn TransformFunc) FormatOption {
	return func(o *formatOptions) {
		o.structuredTransform = fn
	}
}

// WithFormatters sets column formatters applied during table rendering.
// Formatters are only used for table output; json/yaml always use the original body.
func WithFormatters(fmts ColumnFormatters) FormatOption {
	return func(o *formatOptions) {
		o.formatters = fmts
	}
}

// WithColumns sets the column order for table rendering.
// Listed columns are shown first in the given order; any remaining columns follow alphabetically.
func WithColumns(cols ...string) FormatOption {
	return func(o *formatOptions) {
		o.columns = cols
	}
}

// WithDeclaredUnits opts a command into structured-output field rewriting. It is
// the mirror image of WithFormatters: formatters carry unit semantics into the
// table rendering, rewrites carry the same semantics into the yaml / json / --jq
// output that machine callers read. Table output is left untouched.
//
// Commands pass a declaration from internal/unitdecl, which is where the audit
// record of "this payload's time-field units were confirmed" lives. A command
// that passes nothing is not rewritten at all.
func WithDeclaredUnits(rw FieldRewrites) FormatOption {
	return func(o *formatOptions) {
		o.rewrites = rw
	}
}

// FormatOutput renders body according to the output mode (table/yaml/json/jq/compact).
func FormatOutput(body []byte, io *IOStreams, output string, opts ...FormatOption) error {
	var o formatOptions
	for _, opt := range opts {
		opt(&o)
	}

	// Structured output (yaml/json/jq) gets field rewrites; table keeps the
	// original field names and relies on o.formatters instead. Computed lazily so
	// the table path never pays for a rewrite it does not use.
	structured := func() []byte {
		if len(o.rewrites) == 0 {
			return body
		}
		return applyFieldRewrites(body, o.rewrites)
	}

	// The payload structured callers read: rewrites, pagination, unwrapping, then
	// the command's own structured transform.
	structuredBody := func() ([]byte, error) {
		data := unwrapResult(normalizePage(structured()))
		if o.structuredTransform == nil {
			return data, nil
		}
		return o.structuredTransform(data)
	}

	// --jq overrides output mode: apply expression on unwrapped data
	if io.JQExpr != "" {
		data, err := structuredBody()
		if err != nil {
			return err
		}
		result, err := ApplyJQ(data, io.JQExpr)
		if err != nil {
			return err
		}
		if result != "" {
			fmt.Fprintln(io.Out, result)
		}
		return nil
	}

	switch output {
	case "table":
		data := body
		if o.transform != nil {
			var err error
			data, err = o.transform(body)
			if err != nil {
				return err
			}
		}
		if len(o.formatters) > 0 {
			data = applyFormatters(data, o.formatters)
		}
		return FormatTable(data, io, o.columns)
	case "yaml":
		data, err := structuredBody()
		if err != nil {
			return err
		}
		s, err := FormatYAML(data)
		if err != nil {
			return err
		}
		fmt.Fprintln(io.Out, s)
	default:
		if !json.Valid(structured()) {
			fmt.Fprintln(io.Out, string(structured()))
			return nil
		}
		data, err := structuredBody()
		if err != nil {
			return err
		}
		fmt.Fprintln(io.Out, FormatJSON(data, io, output))
	}
	return nil
}

// normalizePage converts the 0-based "page" field in pagination envelopes
// to 1-based, matching the CLI's --page flag convention.
func normalizePage(data []byte) []byte {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}
	pageRaw, ok := raw["page"]
	if !ok {
		return data
	}
	var page int
	if err := json.Unmarshal(pageRaw, &page); err != nil {
		return data
	}
	raw["page"], _ = json.Marshal(page + 1)
	out, err := json.Marshal(raw)
	if err != nil {
		return data
	}
	return out
}

// unwrapResult strips the envelope when the JSON object has "result" as its
// only key (e.g. {"result": {...}} or {"result": [...]}). Multi-key envelopes
// like {"result": [...], "total": 44, "page": 0} are left as-is so callers
// retain access to pagination metadata.
func unwrapResult(data []byte) []byte {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}
	if len(raw) == 1 {
		if inner, ok := raw["result"]; ok {
			return inner
		}
	}
	return data
}

// applyFormatters deserializes JSON, applies column formatters in memory,
// and re-serializes once. Handles both array and single object results.
func applyFormatters(data []byte, fmts ColumnFormatters) []byte {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}

	result, hasEnvelope := raw["result"]
	if !hasEnvelope {
		// No envelope — treat entire object as the item
		applyFormattersToObject(raw, fmts)
		out, err := json.Marshal(raw)
		if err != nil {
			return data
		}
		return out
	}

	switch items := result.(type) {
	case []interface{}:
		for _, item := range items {
			if obj, ok := item.(map[string]interface{}); ok {
				applyFormattersToObject(obj, fmts)
			}
		}
	case map[string]interface{}:
		applyFormattersToObject(items, fmts)
	}

	out, err := json.Marshal(raw)
	if err != nil {
		return data
	}
	return out
}

// applyFormattersToObject applies formatters to matching keys in an object.
// First tries an exact flat key match (e.g. "cpu.usage" as a literal key),
// then falls back to dot-path navigation for nested objects (e.g. "sim" → "tx").
func applyFormattersToObject(obj map[string]interface{}, fmts ColumnFormatters) {
	for col, fn := range fmts {
		// Try exact key first (handles literal dots in key names)
		if v, ok := obj[col]; ok {
			obj[col] = fn(fmt.Sprint(v))
			continue
		}

		// Fall back to dot-path navigation for nested objects
		parts := strings.Split(col, ".")
		if len(parts) < 2 {
			continue
		}
		cur := obj
		for _, p := range parts[:len(parts)-1] {
			nested, ok := cur[p].(map[string]interface{})
			if !ok {
				cur = nil
				break
			}
			cur = nested
		}
		if cur == nil {
			continue
		}
		leaf := parts[len(parts)-1]
		if v, ok := cur[leaf]; ok {
			cur[leaf] = fn(fmt.Sprint(v))
		}
	}
}

// ChainTransforms composes multiple TransformFuncs into one, applying them in order.
func ChainTransforms(fns ...TransformFunc) TransformFunc {
	return func(data []byte) ([]byte, error) {
		var err error
		for _, fn := range fns {
			data, err = fn(data)
			if err != nil {
				return nil, err
			}
		}
		return data, nil
	}
}

// ReverseJSONArray reverses the order of elements in a JSON array.
// It handles both bare arrays ([...]) and enveloped arrays ({"result": [...]}).
func ReverseJSONArray(data []byte) ([]byte, error) {
	// Try {"result": [...]} envelope first
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err == nil {
		if result, ok := envelope["result"]; ok {
			var arr []json.RawMessage
			if err := json.Unmarshal(result, &arr); err == nil {
				slices.Reverse(arr)
				envelope["result"], _ = json.Marshal(arr)
				return json.Marshal(envelope)
			}
		}
	}
	// Try bare array
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil {
		slices.Reverse(arr)
		return json.Marshal(arr)
	}
	return data, nil
}

// ReverseSeriesData reverses the sample order within each series of a
// time-series payload, leaving the envelope, the field list and the series
// grouping as they are. It is the structured-output counterpart to reversing the
// flattened rows a table renders: json / yaml callers keep the series shape they
// already parse, only the samples inside it run newest-first.
//
// Handles both naming conventions FlattenSeries supports (data and values).
// Payloads that are not series-shaped are returned untouched.
func ReverseSeriesData(data []byte) ([]byte, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return data, nil
	}
	seriesRaw, ok := envelope["series"]
	if !ok {
		return data, nil
	}
	var series []map[string]json.RawMessage
	if err := json.Unmarshal(seriesRaw, &series); err != nil {
		return data, nil
	}
	for _, s := range series {
		for _, key := range []string{"data", "values"} {
			raw, ok := s[key]
			if !ok {
				continue
			}
			var rows []json.RawMessage
			if err := json.Unmarshal(raw, &rows); err != nil {
				continue
			}
			slices.Reverse(rows)
			reversed, err := json.Marshal(rows)
			if err != nil {
				continue
			}
			s[key] = reversed
		}
	}
	reordered, err := json.Marshal(series)
	if err != nil {
		return data, nil
	}
	envelope["series"] = reordered
	return json.Marshal(envelope)
}

// FlattenSeries converts a time-series API response (FluxResult) into a flat
// JSON array of objects suitable for FormatTable.
//
// Supports both naming conventions used across backend services:
//   - fields/data (signal, device perf endpoints)
//   - columns/values (FluxResult from uplink, network, data-usage endpoints)
//
// If a series has a non-empty "type" field, it is included in each row.
func FlattenSeries(body []byte) ([]byte, error) {
	var envelope struct {
		Result struct {
			Series []struct {
				Type    string          `json:"type"`
				Fields  []string        `json:"fields"`
				Data    [][]interface{} `json:"data"`
				Columns []string        `json:"columns"`
				Values  [][]interface{} `json:"values"`
			} `json:"series"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parsing series response: %w", err)
	}

	var rows []map[string]interface{}
	for _, s := range envelope.Result.Series {
		cols := s.Fields
		data := s.Data
		if len(cols) == 0 {
			cols = s.Columns
		}
		if len(data) == 0 {
			data = s.Values
		}
		for _, row := range data {
			obj := map[string]interface{}{}
			if s.Type != "" {
				obj["type"] = s.Type
			}
			for i, field := range cols {
				if i < len(row) {
					obj[field] = row[i]
				}
			}
			rows = append(rows, obj)
		}
	}

	if len(rows) == 0 {
		return json.Marshal(map[string]interface{}{"result": []interface{}{}})
	}
	return json.Marshal(map[string]interface{}{"result": rows})
}
