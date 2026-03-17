package api

import (
	"net/url"
	"testing"
)

func TestCleanValues(t *testing.T) {
	tests := []struct {
		name  string
		input url.Values
		want  url.Values
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "all empty values removed",
			input: url.Values{"a": {""}, "b": {""}},
			want:  url.Values{},
		},
		{
			name:  "non-empty values preserved",
			input: url.Values{"type": {"cellular"}, "after": {"2024-01-01"}},
			want:  url.Values{"type": {"cellular"}, "after": {"2024-01-01"}},
		},
		{
			name:  "mixed empty and non-empty",
			input: url.Values{"type": {"cellular"}, "month": {""}, "year": {""}},
			want:  url.Values{"type": {"cellular"}},
		},
		{
			name:  "multi-value with some empty",
			input: url.Values{"groups": {"aaa", "", "bbb"}},
			want:  url.Values{"groups": {"aaa", "bbb"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanValues(tt.input)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("expected %d keys, got %d: %v", len(tt.want), len(got), got)
				return
			}
			for k, wantVals := range tt.want {
				gotVals := got[k]
				if len(gotVals) != len(wantVals) {
					t.Errorf("key %q: expected %v, got %v", k, wantVals, gotVals)
				}
				for i := range wantVals {
					if gotVals[i] != wantVals[i] {
						t.Errorf("key %q[%d]: expected %q, got %q", k, i, wantVals[i], gotVals[i])
					}
				}
			}
		})
	}
}
