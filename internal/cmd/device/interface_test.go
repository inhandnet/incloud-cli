package device

import (
	"encoding/json"
	"testing"
)

func flattenRows(t *testing.T, body string) []map[string]any {
	t.Helper()
	out, err := flattenInterfaces([]byte(body))
	if err != nil {
		t.Fatalf("flattenInterfaces: %v", err)
	}
	var wrapped struct {
		Result []map[string]any `json:"result"`
	}
	if err := json.Unmarshal(out, &wrapped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return wrapped.Result
}

func TestFlattenInterfacesKeepsMac(t *testing.T) {
	rows := flattenRows(t, `{"result":{"wan":[{"name":"wan1","mac":"00:18:05:36:59:DE"}]}}`)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0]["mac"] != "00:18:05:36:59:DE" {
		t.Errorf("expected mac preserved, got %v", rows[0]["mac"])
	}
}

// The table renderer derives columns from the first row, so a mac reported only by
// a later interface must still be backfilled as an empty placeholder on earlier rows.
func TestFlattenInterfacesBackfillsSparseMac(t *testing.T) {
	rows := flattenRows(t, `{"result":{"cellular":[{"name":"cellular1"}],"wan":[{"name":"wan1","mac":"00:18:05:36:59:DE"}]}}`)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	mac, ok := rows[0]["mac"]
	if !ok {
		t.Fatal("expected mac placeholder on first row")
	}
	if mac != "" {
		t.Errorf("expected empty placeholder, got %v", mac)
	}
	if rows[1]["mac"] != "00:18:05:36:59:DE" {
		t.Errorf("expected mac on wan row, got %v", rows[1]["mac"])
	}
}

func TestFlattenInterfacesOmitsMacWhenAbsentEverywhere(t *testing.T) {
	rows := flattenRows(t, `{"result":{"wan":[{"name":"wan1"}],"lan":[{"name":"lan2"}]}}`)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for i, row := range rows {
		if _, ok := row["mac"]; ok {
			t.Errorf("row %d should not carry a mac key", i)
		}
	}
}
