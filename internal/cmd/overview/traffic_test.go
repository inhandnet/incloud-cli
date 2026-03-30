package overview

import (
	"encoding/json"
	"testing"
)

var testOverviewRaw = json.RawMessage(`{
  "series": [
    {
      "type": "cellular",
      "fields": ["ts", "tx", "rx", "total"],
      "data": [
        ["2024-01-01T00:00:00Z", 100, 200, 300],
        ["2024-01-02T00:00:00Z", 150, 250, 400]
      ]
    },
    {
      "type": "wired",
      "fields": ["ts", "tx", "rx", "total"],
      "data": [
        ["2024-01-01T00:00:00Z", 50, 80, 130]
      ]
    }
  ]
}`)

func TestBuildTrafficSummary(t *testing.T) {
	rows := buildTrafficSummary(testOverviewRaw)

	if len(rows) != 2 {
		t.Fatalf("expected 2 summary rows, got %d", len(rows))
	}

	cellular := rows[0]
	if cellular["type"] != "cellular" {
		t.Errorf("expected type=cellular, got %v", cellular["type"])
	}
	if cellular["tx"] != float64(250) {
		t.Errorf("expected cellular tx=250, got %v", cellular["tx"])
	}
	if cellular["rx"] != float64(450) {
		t.Errorf("expected cellular rx=450, got %v", cellular["rx"])
	}
	if cellular["total"] != float64(700) {
		t.Errorf("expected cellular total=700, got %v", cellular["total"])
	}

	wired := rows[1]
	if wired["type"] != "wired" {
		t.Errorf("expected type=wired, got %v", wired["type"])
	}
	if wired["total"] != float64(130) {
		t.Errorf("expected wired total=130, got %v", wired["total"])
	}
}

func TestBuildTrafficTrend(t *testing.T) {
	rows := buildTrafficTrend(testOverviewRaw)

	// 2 cellular rows + 1 wired row
	if len(rows) != 3 {
		t.Fatalf("expected 3 trend rows, got %d", len(rows))
	}

	// Each row must have type + all series fields
	for _, row := range rows {
		if _, ok := row["type"]; !ok {
			t.Error("trend row missing 'type' field")
		}
		if _, ok := row["ts"]; !ok {
			t.Error("trend row missing 'ts' field")
		}
		if _, ok := row["tx"]; !ok {
			t.Error("trend row missing 'tx' field")
		}
	}

	if rows[0]["type"] != "cellular" {
		t.Errorf("expected first row type=cellular, got %v", rows[0]["type"])
	}
	if rows[2]["type"] != "wired" {
		t.Errorf("expected third row type=wired, got %v", rows[2]["type"])
	}
}

func TestBuildTrafficSummary_Empty(t *testing.T) {
	rows := buildTrafficSummary(json.RawMessage(`{"series":[]}`))
	if len(rows) != 0 {
		t.Errorf("expected empty summary for empty series, got %d rows", len(rows))
	}
}

func TestBuildTrafficSummary_Invalid(t *testing.T) {
	rows := buildTrafficSummary(json.RawMessage(`invalid`))
	if rows != nil {
		t.Errorf("expected nil for invalid input, got %v", rows)
	}
}
