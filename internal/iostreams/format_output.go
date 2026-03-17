package iostreams

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

// applyFormatters rewrites JSON values in-place for columns that have formatters.
// It handles both array (result is array of objects) and single object cases.
func applyFormatters(data []byte, fmts ColumnFormatters) []byte {
	raw := gjson.ParseBytes(data)

	items := raw.Get("result")
	if !items.Exists() {
		items = raw
	}

	switch {
	case items.IsArray():
		arr := items.Array()
		for i, item := range arr {
			if !item.IsObject() {
				continue
			}
			for col, fn := range fmts {
				v := item.Get(col)
				if !v.Exists() {
					continue
				}
				path := fmt.Sprintf("result.%d.%s", i, col)
				formatted := fn(formatResult(&v))
				var err error
				data, err = sjson.SetBytes(data, path, formatted)
				if err != nil {
					continue
				}
			}
		}
	case items.IsObject():
		for col, fn := range fmts {
			v := items.Get(col)
			if !v.Exists() {
				continue
			}
			path := "result." + col
			if !raw.Get("result").Exists() {
				path = col
			}
			formatted := fn(formatResult(&v))
			var err error
			data, err = sjson.SetBytes(data, path, formatted)
			if err != nil {
				continue
			}
		}
	}

	return data
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
