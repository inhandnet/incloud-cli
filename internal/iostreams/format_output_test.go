package iostreams

import (
	"encoding/json"
	"testing"
)

func TestReverseJSONArray_BareArray(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "three elements",
			data: `[{"t":1},{"t":2},{"t":3}]`,
			want: `[{"t":3},{"t":2},{"t":1}]`,
		},
		{
			name: "single element",
			data: `[{"t":1}]`,
			want: `[{"t":1}]`,
		},
		{
			name: "empty array",
			data: `[]`,
			want: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReverseJSONArray([]byte(tt.data))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(t, got, []byte(tt.want)) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestReverseJSONArray_Envelope(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "result array reversed",
			data: `{"result":[{"t":1},{"t":2},{"t":3}]}`,
			want: `{"result":[{"t":3},{"t":2},{"t":1}]}`,
		},
		{
			name: "empty result array",
			data: `{"result":[]}`,
			want: `{"result":[]}`,
		},
		{
			name: "result is not array — unchanged",
			data: `{"result":{"key":"value"}}`,
			want: `{"result":{"key":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReverseJSONArray([]byte(tt.data))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(t, got, []byte(tt.want)) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestReverseJSONArray_NonArrayInput(t *testing.T) {
	input := `"just a string"`
	got, err := ReverseJSONArray([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != input {
		t.Errorf("got %s, want %s", got, input)
	}
}

func TestChainTransforms(t *testing.T) {
	addWrapper := func(data []byte) ([]byte, error) {
		return json.Marshal(map[string]json.RawMessage{"result": data})
	}

	chain := ChainTransforms(addWrapper, ReverseJSONArray)

	input := `[{"t":1},{"t":2},{"t":3}]`
	got, err := chain([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"result":[{"t":3},{"t":2},{"t":1}]}`
	if !jsonEqual(t, got, []byte(want)) {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestChainTransforms_ErrorPropagation(t *testing.T) {
	fail := func([]byte) ([]byte, error) {
		var v any
		return nil, json.Unmarshal([]byte("invalid"), &v)
	}
	noop := func(data []byte) ([]byte, error) { return data, nil }

	chain := ChainTransforms(fail, noop)
	_, err := chain([]byte(`[]`))
	if err == nil {
		t.Error("expected error from first transform to propagate")
	}
}

// jsonEqual compares two JSON byte slices for semantic equality.
func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var va, vb interface{}
	if err := json.Unmarshal(a, &va); err != nil {
		t.Fatalf("failed to unmarshal a: %v", err)
	}
	if err := json.Unmarshal(b, &vb); err != nil {
		t.Fatalf("failed to unmarshal b: %v", err)
	}
	aj, _ := json.Marshal(va)
	bj, _ := json.Marshal(vb)
	return string(aj) == string(bj)
}
