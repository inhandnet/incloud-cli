package iostreams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

const (
	// TrafficHumanUnit is the single unit used by structured traffic output.
	TrafficHumanUnit = "GiB"
	// TrafficUnitSystem describes the byte base used by TrafficHumanUnit.
	TrafficUnitSystem = "IEC (1 GiB = 1024^3 B)"
	// TrafficBytesPerUnit is the number of bytes in one TrafficHumanUnit.
	TrafficBytesPerUnit int64 = 1 << 30
)

var trafficRawFields = []struct {
	raw   string
	human string
}{
	{raw: "tx", human: "txHuman"},
	{raw: "rx", human: "rxHuman"},
	{raw: "total", human: "totalHuman"},
}

// AddTrafficHumanFields adds deterministic, IEC-based display fields to a
// traffic payload. Raw byte fields are retained for exact calculations.
//
// The transform handles both row-shaped objects and columnar time-series
// objects (fields/data and columns/values). It is intended for structured
// output only; table output continues to use the existing column formatters.
func AddTrafficHumanFields(body []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()

	var payload any
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}

	normalizeTrafficValue(payload)
	return json.Marshal(payload)
}

func normalizeTrafficValue(value any) {
	switch v := value.(type) {
	case map[string]any:
		for _, child := range v {
			normalizeTrafficValue(child)
		}
		normalizeTrafficSeries(v)
		addTrafficObjectFields(v)
		if _, ok := v["summary"]; ok {
			setTrafficMetadata(v)
		}
	case []any:
		for _, child := range v {
			normalizeTrafficValue(child)
		}
	}
}

func addTrafficObjectFields(obj map[string]any) {
	var (
		values  = make(map[string]*big.Int, len(trafficRawFields))
		matched bool
	)

	for _, field := range trafficRawFields {
		raw, ok := obj[field.raw]
		if !ok {
			continue
		}
		bytesValue, ok := trafficBytesValue(raw)
		if !ok {
			continue
		}
		obj[field.human] = formatTrafficGiB(bytesValue)
		values[field.raw] = bytesValue
		matched = true
	}

	if !matched {
		return
	}

	setTrafficMetadata(obj)
	if len(values) == len(trafficRawFields) {
		var parts big.Int
		parts.Add(values["rx"], values["tx"])
		obj["trafficReconciled"] = parts.Cmp(values["total"]) == 0
	}
}

func normalizeTrafficSeries(obj map[string]any) {
	columnKey, rowKey := seriesKeys(obj)
	if columnKey == "" {
		return
	}

	columns, _ := obj[columnKey].([]any)
	rows, _ := obj[rowKey].([]any)
	indexes := make(map[string]int, len(columns))
	for i, column := range columns {
		name, ok := column.(string)
		if ok {
			indexes[name] = i
		}
	}

	humanIndexes := make(map[string]int, len(trafficRawFields))
	for _, field := range trafficRawFields {
		if _, ok := indexes[field.raw]; !ok {
			continue
		}
		index, ok := indexes[field.human]
		if !ok {
			index = len(columns)
			columns = append(columns, field.human)
			indexes[field.human] = index
		}
		humanIndexes[field.raw] = index
	}

	if len(humanIndexes) == 0 {
		return
	}

	reconciledIndex := -1
	if len(humanIndexes) == len(trafficRawFields) {
		if index, ok := indexes["trafficReconciled"]; ok {
			reconciledIndex = index
		} else {
			reconciledIndex = len(columns)
			columns = append(columns, "trafficReconciled")
			indexes["trafficReconciled"] = reconciledIndex
		}
	}

	setTrafficMetadata(obj)
	for rowIndex, row := range rows {
		cells, ok := row.([]any)
		if !ok {
			continue
		}
		for len(cells) < len(columns) {
			cells = append(cells, nil)
		}
		for _, field := range trafficRawFields {
			rawIndex, rawOK := indexes[field.raw]
			humanIndex, humanOK := humanIndexes[field.raw]
			if !rawOK || !humanOK || rawIndex >= len(cells) {
				continue
			}
			if bytesValue, ok := trafficBytesValue(cells[rawIndex]); ok {
				cells[humanIndex] = formatTrafficGiB(bytesValue)
			}
		}
		if reconciledIndex >= 0 {
			values := make(map[string]*big.Int, len(trafficRawFields))
			for _, field := range trafficRawFields {
				rawIndex := indexes[field.raw]
				if bytesValue, ok := trafficBytesValue(cells[rawIndex]); ok {
					values[field.raw] = bytesValue
				}
			}
			reconciled := false
			if len(values) == len(trafficRawFields) {
				var parts big.Int
				parts.Add(values["rx"], values["tx"])
				reconciled = parts.Cmp(values["total"]) == 0
			}
			cells[reconciledIndex] = reconciled
		}
		rows[rowIndex] = cells
	}

	obj[columnKey] = columns
	obj[rowKey] = rows
}

func setTrafficMetadata(obj map[string]any) {
	obj["trafficUnit"] = TrafficHumanUnit
	obj["trafficUnitSystem"] = TrafficUnitSystem
	obj["trafficBytesPerUnit"] = TrafficBytesPerUnit
}

func trafficBytesValue(value any) (*big.Int, bool) {
	text, ok := trafficNumberString(value)
	if !ok {
		return nil, false
	}

	rational, ok := new(big.Rat).SetString(text)
	if !ok || rational.Denom().Cmp(big.NewInt(1)) != 0 {
		return nil, false
	}
	return new(big.Int).Set(rational.Num()), true
}

func trafficNumberString(value any) (string, bool) {
	switch v := value.(type) {
	case json.Number:
		return v.String(), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case int:
		return strconv.Itoa(v), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	default:
		return "", false
	}
}

func formatTrafficGiB(bytesValue *big.Int) string {
	value := new(big.Rat).SetInt(bytesValue)
	value.Quo(value, new(big.Rat).SetInt64(TrafficBytesPerUnit))
	return fmt.Sprintf("%s %s", value.FloatString(2), TrafficHumanUnit)
}
