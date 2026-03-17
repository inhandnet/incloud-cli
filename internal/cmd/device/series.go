package device

import (
	"encoding/json"
	"fmt"
)

// flattenSeries converts a time-series API response into a flat JSON array
// suitable for FormatTable. Supports both naming conventions:
//   - fields/data (signal, perf endpoints)
//   - columns/values (FluxResult from uplink perf-trend, etc.)
//
// Used by endpoints that have no "type" per series.
func flattenSeries(body []byte) ([]byte, error) {
	return flattenSeriesImpl(body, false)
}

// flattenSeriesWithType is like flattenSeries but includes the series "type" field
// in each row. Used by signal endpoints where each series has a type (e.g. "4G").
func flattenSeriesWithType(body []byte) ([]byte, error) {
	return flattenSeriesImpl(body, true)
}

func flattenSeriesImpl(body []byte, includeType bool) ([]byte, error) {
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
			if includeType {
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
