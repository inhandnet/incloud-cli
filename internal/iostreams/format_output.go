package iostreams

import (
	"encoding/json"
	"fmt"
)

// TransformFunc converts raw API response bytes into a format suitable for FormatTable.
type TransformFunc func([]byte) ([]byte, error)

// FormatOption configures FormatOutput behavior.
type FormatOption func(*formatOptions)

type formatOptions struct {
	transform  TransformFunc
	formatters ColumnFormatters
}

// WithTransform sets a transform function applied before table rendering.
// The transform is only used for table output; json/yaml always use the original body.
func WithTransform(fn TransformFunc) FormatOption {
	return func(o *formatOptions) {
		o.transform = fn
	}
}

// WithFormatters sets column formatters applied during table rendering.
// Formatters are only used for table output; json/yaml always use the original body.
func WithFormatters(fmts ColumnFormatters) FormatOption {
	return func(o *formatOptions) {
		o.formatters = fmts
	}
}

// FormatOutput renders body according to the output mode (table/yaml/json/jq/compact).
// fields controls which columns to show in table mode (empty = all).
func FormatOutput(body []byte, io *IOStreams, output string, fields []string, opts ...FormatOption) error {
	var o formatOptions
	for _, opt := range opts {
		opt(&o)
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
		return FormatTable(data, io, fields)
	case "yaml":
		s, err := FormatYAML(body)
		if err != nil {
			return err
		}
		fmt.Fprintln(io.Out, s)
	default:
		if json.Valid(body) {
			fmt.Fprintln(io.Out, FormatJSON(body, io, output))
		} else {
			fmt.Fprintln(io.Out, string(body))
		}
	}
	return nil
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

// applyFormattersToObject applies formatters to matching keys in a flat object.
func applyFormattersToObject(obj map[string]interface{}, fmts ColumnFormatters) {
	for col, fn := range fmts {
		v, ok := obj[col]
		if !ok {
			continue
		}
		obj[col] = fn(fmt.Sprint(v))
	}
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
