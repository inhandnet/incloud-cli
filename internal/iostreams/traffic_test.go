package iostreams

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAddTrafficHumanFields_Rows(t *testing.T) {
	input := []byte(`{
		"summary":[{"tx":312377884,"rx":3269296004,"total":3581673888}],
		"trend":[{"type":"cellular","tx":285132646,"rx":936144800,"total":1221277446}]
	}`)

	got, err := AddTrafficHumanFields(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload struct {
		TrafficUnit       string                   `json:"trafficUnit"`
		TrafficUnitSystem string                   `json:"trafficUnitSystem"`
		Summary           []map[string]interface{} `json:"summary"`
		Trend             []map[string]interface{} `json:"trend"`
	}
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("invalid output: %v", err)
	}

	if payload.TrafficUnit != TrafficHumanUnit {
		t.Errorf("trafficUnit = %q, want %q", payload.TrafficUnit, TrafficHumanUnit)
	}
	if payload.TrafficUnitSystem != TrafficUnitSystem {
		t.Errorf("trafficUnitSystem = %q, want %q", payload.TrafficUnitSystem, TrafficUnitSystem)
	}
	if len(payload.Summary) != 1 || len(payload.Trend) != 1 {
		t.Fatalf("unexpected row counts: summary=%d trend=%d", len(payload.Summary), len(payload.Trend))
	}

	row := payload.Summary[0]
	if row["tx"] != float64(312377884) {
		t.Errorf("raw tx was changed: %v", row["tx"])
	}
	if row["txHuman"] != "0.29 GiB" {
		t.Errorf("txHuman = %v, want 0.29 GiB", row["txHuman"])
	}
	if row["rxHuman"] != "3.04 GiB" {
		t.Errorf("rxHuman = %v, want 3.04 GiB", row["rxHuman"])
	}
	if row["totalHuman"] != "3.34 GiB" {
		t.Errorf("totalHuman = %v, want 3.34 GiB", row["totalHuman"])
	}
	if row["trafficReconciled"] != true {
		t.Errorf("trafficReconciled = %v, want true", row["trafficReconciled"])
	}

	trend := payload.Trend[0]
	if trend["totalHuman"] != "1.14 GiB" {
		t.Errorf("trend totalHuman = %v, want 1.14 GiB", trend["totalHuman"])
	}
}

func TestAddTrafficHumanFields_ColumnarSeries(t *testing.T) {
	input := []byte(`{
		"series":[{
			"fields":["time","tx","rx","total"],
			"data":[["2026-08-06T00:00:00Z",312377884,3269296004,3581673888]]
		}]
	}`)

	got, err := AddTrafficHumanFields(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload struct {
		Series []struct {
			Fields []string        `json:"fields"`
			Data   [][]interface{} `json:"data"`
		} `json:"series"`
	}
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("invalid output: %v", err)
	}
	if len(payload.Series) != 1 {
		t.Fatalf("expected one series, got %d", len(payload.Series))
	}

	series := payload.Series[0]
	wantFields := []string{"time", "tx", "rx", "total", "txHuman", "rxHuman", "totalHuman", "trafficReconciled"}
	if strings.Join(series.Fields, "|") != strings.Join(wantFields, "|") {
		t.Fatalf("fields = %v, want %v", series.Fields, wantFields)
	}
	if len(series.Data) != 1 || len(series.Data[0]) != len(wantFields) {
		t.Fatalf("unexpected data shape: %#v", series.Data)
	}
	row := series.Data[0]
	if row[4] != "0.29 GiB" || row[5] != "3.04 GiB" || row[6] != "3.34 GiB" || row[7] != true {
		t.Errorf("human cells = %v, want [0.29 GiB 3.04 GiB 3.34 GiB]", row[4:])
	}
}

func TestAddTrafficHumanFields_InvalidJSON(t *testing.T) {
	if _, err := AddTrafficHumanFields([]byte("not-json")); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestAddTrafficHumanFields_IM3173Samples(t *testing.T) {
	input := []byte(`{"rows":[
		{"date":"08-06","tx":312377884,"rx":3269296004,"total":3581673888},
		{"date":"08-07","tx":285132646,"rx":936144800,"total":1221277446},
		{"date":"08-13","tx":60405546,"rx":720972738,"total":781378284},
		{"date":"summary","tx":1299144753,"rx":11290795623,"total":12589940376}
	]}`)

	got, err := AddTrafficHumanFields(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var payload struct {
		Rows []map[string]interface{} `json:"rows"`
	}
	if err := json.Unmarshal(got, &payload); err != nil {
		t.Fatalf("invalid output: %v", err)
	}

	want := []struct {
		date          string
		tx, rx, total string
	}{
		{date: "08-06", tx: "0.29 GiB", rx: "3.04 GiB", total: "3.34 GiB"},
		{date: "08-07", tx: "0.27 GiB", rx: "0.87 GiB", total: "1.14 GiB"},
		{date: "08-13", tx: "0.06 GiB", rx: "0.67 GiB", total: "0.73 GiB"},
		{date: "summary", tx: "1.21 GiB", rx: "10.52 GiB", total: "11.73 GiB"},
	}
	if len(payload.Rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(payload.Rows), len(want))
	}
	for i, expected := range want {
		row := payload.Rows[i]
		if row["date"] != expected.date || row["txHuman"] != expected.tx ||
			row["rxHuman"] != expected.rx || row["totalHuman"] != expected.total {
			t.Errorf("row %d = %v, want %s: %s/%s/%s", i, row, expected.date, expected.tx, expected.rx, expected.total)
		}
		if row["trafficReconciled"] != true {
			t.Errorf("row %d is not reconciled: %v", i, row["trafficReconciled"])
		}
	}
}

func TestTrafficHumanFieldsStructuredOnly(t *testing.T) {
	body := []byte(`{"result":[{"tx":312377884,"rx":3269296004,"total":3581673888}]}`)

	jsonIO, jsonOut := newBufferIO()
	if err := FormatOutput(body, jsonIO, "json", WithStructuredTransform(AddTrafficHumanFields)); err != nil {
		t.Fatalf("json output: %v", err)
	}
	if !strings.Contains(jsonOut.String(), `"totalHuman"`) || !strings.Contains(jsonOut.String(), `3.34 GiB`) {
		t.Errorf("structured output missing totalHuman: %s", jsonOut.String())
	}

	tableIO, tableOut := newBufferIO()
	if err := FormatOutput(body, tableIO, "table", WithStructuredTransform(AddTrafficHumanFields)); err != nil {
		t.Fatalf("table output: %v", err)
	}
	if strings.Contains(tableOut.String(), "totalHuman") {
		t.Errorf("table output unexpectedly contains structured human field: %s", tableOut.String())
	}
}
