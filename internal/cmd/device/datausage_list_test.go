package device

import (
	"encoding/json"
	"testing"
)

func TestFlattenDatausageDetails(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, result []map[string]interface{})
	}{
		{
			name: "strips time and preserves nested structure",
			input: `{"result":[{
				"deviceId":"aaa",
				"sim":{"time":"2026-03-01T00:00:00Z","tx":100,"rx":200,"total":300},
				"esim":{"time":"2026-03-01T00:00:00Z","tx":50,"rx":60,"total":110}
			}]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				if len(result) != 1 {
					t.Fatalf("expected 1 device, got %d", len(result))
				}
				dev := result[0]
				if dev["deviceId"] != "aaa" {
					t.Errorf("expected deviceId=aaa, got %v", dev["deviceId"])
				}
				// Check sim: time stripped, data preserved
				sim, ok := dev["sim"].(map[string]interface{})
				if !ok {
					t.Fatal("sim should be a nested object")
				}
				if _, hasTime := sim["time"]; hasTime {
					t.Error("time field should be stripped from sim")
				}
				if sim["tx"] != float64(100) {
					t.Errorf("expected sim.tx=100, got %v", sim["tx"])
				}
				// Check esim: time also stripped
				esim, ok := dev["esim"].(map[string]interface{})
				if !ok {
					t.Fatal("esim should be a nested object")
				}
				if _, hasTime := esim["time"]; hasTime {
					t.Error("time field should be stripped from esim")
				}
				if esim["total"] != float64(110) {
					t.Errorf("expected esim.total=110, got %v", esim["total"])
				}
			},
		},
		{
			name: "sorts by deviceId",
			input: `{"result":[
				{"deviceId":"ccc","sim":{"tx":1,"rx":2,"total":3}},
				{"deviceId":"aaa","sim":{"tx":4,"rx":5,"total":9}},
				{"deviceId":"bbb","sim":{"tx":7,"rx":8,"total":15}}
			]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				if len(result) != 3 {
					t.Fatalf("expected 3 devices, got %d", len(result))
				}
				ids := make([]string, len(result))
				for i, d := range result {
					ids[i], _ = d["deviceId"].(string)
				}
				if ids[0] != "aaa" || ids[1] != "bbb" || ids[2] != "ccc" {
					t.Errorf("expected sorted order [aaa,bbb,ccc], got %v", ids)
				}
			},
		},
		{
			name:  "empty result",
			input: `{"result":[]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				if len(result) != 0 {
					t.Errorf("expected empty result, got %d items", len(result))
				}
			},
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
		{
			name: "device with only scalar fields (no nested objects)",
			input: `{"result":[{
				"deviceId":"xxx"
			}]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				if len(result) != 1 {
					t.Fatalf("expected 1 device, got %d", len(result))
				}
				if result[0]["deviceId"] != "xxx" {
					t.Errorf("expected deviceId=xxx, got %v", result[0]["deviceId"])
				}
			},
		},
		{
			name: "nested object without time key is unchanged",
			input: `{"result":[{
				"deviceId":"aaa",
				"sim":{"tx":100,"rx":200,"total":300}
			}]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				sim, _ := result[0]["sim"].(map[string]interface{})
				if sim["tx"] != float64(100) || sim["rx"] != float64(200) || sim["total"] != float64(300) {
					t.Errorf("expected sim fields preserved, got %v", sim)
				}
			},
		},
		{
			name: "mixed scalar and nested values on same device",
			input: `{"result":[{
				"deviceId":"aaa",
				"sim":{"time":"2026-03-01T00:00:00Z","tx":100,"rx":200,"total":300},
				"extraField":42
			}]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				dev := result[0]
				// Scalar field preserved as-is
				if dev["extraField"] != float64(42) {
					t.Errorf("expected extraField=42, got %v", dev["extraField"])
				}
				// Nested sim still processed
				sim, _ := dev["sim"].(map[string]interface{})
				if _, hasTime := sim["time"]; hasTime {
					t.Error("time should be stripped from sim")
				}
			},
		},
		{
			name: "missing deviceId sorts to front",
			input: `{"result":[
				{"deviceId":"bbb","sim":{"tx":1}},
				{"sim":{"tx":2}}
			]}`,
			checkFunc: func(t *testing.T, result []map[string]interface{}) {
				// Missing deviceId type-asserts to "" which sorts before "bbb"
				first, _ := result[0]["deviceId"].(string)
				if first != "" {
					t.Errorf("expected empty deviceId first, got %q", first)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenDatausageDetails([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var envelope struct {
				Result []map[string]interface{} `json:"result"`
			}
			if err := json.Unmarshal(got, &envelope); err != nil {
				t.Fatalf("failed to parse output: %v", err)
			}
			tt.checkFunc(t, envelope.Result)
		})
	}
}
