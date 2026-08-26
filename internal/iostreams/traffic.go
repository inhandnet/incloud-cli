package iostreams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

const (
	// TrafficHumanUnit is the default unit kept for compatibility with callers
	// that used the original fixed-unit implementation. Structured output now
	// selects the unit from trafficUnits based on the largest value.
	TrafficHumanUnit = "GiB"
	// TrafficUnitSystem describes the default unit's byte base.
	TrafficUnitSystem = "IEC (1 GiB = 1024^3 B)"
	// TrafficBytesPerUnit is the number of bytes in the default unit.
	TrafficBytesPerUnit int64 = 1 << 30
)

type trafficUnit struct {
	name         string
	bytesPerUnit int64
	power        uint
}

var trafficUnits = []trafficUnit{
	{name: "B", bytesPerUnit: 1},
	{name: "KiB", bytesPerUnit: 1 << 10, power: 1},
	{name: "MiB", bytesPerUnit: 1 << 20, power: 2},
	{name: "GiB", bytesPerUnit: 1 << 30, power: 3},
	{name: "TiB", bytesPerUnit: 1 << 40, power: 4},
	{name: "PiB", bytesPerUnit: 1 << 50, power: 5},
	{name: "EiB", bytesPerUnit: 1 << 60, power: 6},
}

const trafficHumanPrecision = 3

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

	unit := selectTrafficUnit(payload)
	normalizeTrafficValue(payload, unit)
	return json.Marshal(payload)
}

func selectTrafficUnit(value any) trafficUnit {
	maxBytes := new(big.Int)
	collectTrafficBytes(value, maxBytes)

	selected := trafficUnits[0]
	for _, candidate := range trafficUnits {
		if maxBytes.Cmp(big.NewInt(candidate.bytesPerUnit)) >= 0 {
			selected = candidate
		}
	}
	return selected
}

func collectTrafficBytes(value any, maxBytes *big.Int) {
	switch v := value.(type) {
	case map[string]any:
		for _, field := range trafficRawFields {
			if raw, ok := v[field.raw]; ok {
				updateMaxTrafficBytes(raw, maxBytes)
			}
		}

		// Columnar series keep field names in a separate columns/fields array,
		// so inspect those rows explicitly before walking nested values.
		if columnKey, rowKey := seriesKeys(v); columnKey != "" {
			columns, _ := v[columnKey].([]any)
			rows, _ := v[rowKey].([]any)
			indexes := make(map[string]int, len(columns))
			for i, column := range columns {
				if name, ok := column.(string); ok {
					indexes[name] = i
				}
			}
			for _, field := range trafficRawFields {
				index, ok := indexes[field.raw]
				if !ok {
					continue
				}
				for _, row := range rows {
					cells, ok := row.([]any)
					if ok && index < len(cells) {
						updateMaxTrafficBytes(cells[index], maxBytes)
					}
				}
			}
		}

		for _, child := range v {
			collectTrafficBytes(child, maxBytes)
		}
	case []any:
		for _, child := range v {
			collectTrafficBytes(child, maxBytes)
		}
	}
}

func updateMaxTrafficBytes(value any, maxBytes *big.Int) {
	bytesValue, ok := trafficBytesValue(value)
	if ok && bytesValue.Cmp(maxBytes) > 0 {
		maxBytes.Set(bytesValue)
	}
}

func normalizeTrafficValue(value any, unit trafficUnit) {
	switch v := value.(type) {
	case map[string]any:
		for _, child := range v {
			normalizeTrafficValue(child, unit)
		}
		normalizeTrafficSeries(v, unit)
		addTrafficObjectFields(v, unit)
		if _, ok := v["summary"]; ok {
			setTrafficMetadata(v, unit)
		}
	case []any:
		for _, child := range v {
			normalizeTrafficValue(child, unit)
		}
	}
}

func addTrafficObjectFields(obj map[string]any, unit trafficUnit) {
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
		obj[field.human] = formatTraffic(bytesValue, unit)
		values[field.raw] = bytesValue
		matched = true
	}

	if !matched {
		return
	}

	setTrafficMetadata(obj, unit)
	if len(values) == len(trafficRawFields) {
		reconciled := trafficValuesReconciled(values)
		obj["trafficReconciled"] = reconciled
		if reconciled {
			// Keep the displayed row internally additive: totalHuman is the
			// sum of the already-rounded RX/TX values, not an independent
			// rounding of the raw total.
			obj["totalHuman"] = formatTrafficParts(values["rx"], values["tx"], unit)
		}
	}
}

func normalizeTrafficSeries(obj map[string]any, unit trafficUnit) {
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

	setTrafficMetadata(obj, unit)
	for rowIndex, row := range rows {
		cells, ok := row.([]any)
		if !ok {
			continue
		}
		for len(cells) < len(columns) {
			cells = append(cells, nil)
		}
		values := make(map[string]*big.Int, len(trafficRawFields))
		for _, field := range trafficRawFields {
			rawIndex, rawOK := indexes[field.raw]
			humanIndex, humanOK := humanIndexes[field.raw]
			if !rawOK || !humanOK || rawIndex >= len(cells) {
				continue
			}
			if bytesValue, ok := trafficBytesValue(cells[rawIndex]); ok {
				cells[humanIndex] = formatTraffic(bytesValue, unit)
				values[field.raw] = bytesValue
			}
		}
		if reconciledIndex >= 0 {
			reconciled := trafficValuesReconciled(values)
			if reconciled {
				cells[humanIndexes["total"]] = formatTrafficParts(
					values["rx"], values["tx"], unit,
				)
			}
			cells[reconciledIndex] = reconciled
		}
		rows[rowIndex] = cells
	}

	obj[columnKey] = columns
	obj[rowKey] = rows
}

func setTrafficMetadata(obj map[string]any, unit trafficUnit) {
	obj["trafficUnit"] = unit.name
	obj["trafficUnitSystem"] = unit.system()
	obj["trafficBytesPerUnit"] = unit.bytesPerUnit
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

func trafficValuesReconciled(values map[string]*big.Int) bool {
	if len(values) != len(trafficRawFields) {
		return false
	}

	var parts big.Int
	parts.Add(values["rx"], values["tx"])
	return parts.Cmp(values["total"]) == 0
}

func (u trafficUnit) system() string {
	if u.power == 0 {
		return "IEC (bytes)"
	}
	return fmt.Sprintf("IEC (1 %s = 1024^%d B)", u.name, u.power)
}

func formatTraffic(bytesValue *big.Int, unit trafficUnit) string {
	return formatTrafficRat(roundedTrafficValue(bytesValue, unit), unit)
}

func formatTrafficParts(rx, tx *big.Int, unit trafficUnit) string {
	value := new(big.Rat).Add(
		roundedTrafficValue(rx, unit),
		roundedTrafficValue(tx, unit),
	)
	return formatTrafficRat(value, unit)
}

func roundedTrafficValue(bytesValue *big.Int, unit trafficUnit) *big.Rat {
	value := new(big.Rat).SetInt(bytesValue)
	value.Quo(value, new(big.Rat).SetInt64(unit.bytesPerUnit))
	rounded, ok := new(big.Rat).SetString(value.FloatString(trafficHumanPrecision))
	if !ok {
		return value
	}
	return rounded
}

func formatTrafficRat(value *big.Rat, unit trafficUnit) string {
	return fmt.Sprintf("%s %s", value.FloatString(trafficHumanPrecision), unit.name)
}
