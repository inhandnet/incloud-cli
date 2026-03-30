package api

import "testing"

func TestResultIDName(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantID   string
		wantName string
	}{
		{
			name:     "standard response",
			body:     `{"result":{"_id":"abc123","name":"My Resource"}}`,
			wantID:   "abc123",
			wantName: "My Resource",
		},
		{
			name:     "extra fields ignored",
			body:     `{"result":{"_id":"id1","name":"r1","status":"active"}}`,
			wantID:   "id1",
			wantName: "r1",
		},
		{
			name:     "missing name",
			body:     `{"result":{"_id":"id1"}}`,
			wantID:   "id1",
			wantName: "",
		},
		{
			name:     "missing id",
			body:     `{"result":{"name":"r1"}}`,
			wantID:   "",
			wantName: "r1",
		},
		{
			name:     "empty result",
			body:     `{"result":{}}`,
			wantID:   "",
			wantName: "",
		},
		{
			name:     "no result key",
			body:     `{"data":{"_id":"id1","name":"r1"}}`,
			wantID:   "",
			wantName: "",
		},
		{
			name:     "invalid json",
			body:     `not json`,
			wantID:   "",
			wantName: "",
		},
		{
			name:     "empty body",
			body:     ``,
			wantID:   "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, name := ResultIDName([]byte(tt.body))
			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}
