package iostreams

import (
	"encoding/json"
	"strings"
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

func TestReverseSeriesData(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "samples reversed, fields and envelope kept",
			data: `{"series":[{"fields":["time","rsrp"],"data":[["t1",-100],["t2",-90],["t3",-80]]}]}`,
			want: `{"series":[{"fields":["time","rsrp"],"data":[["t3",-80],["t2",-90],["t1",-100]]}]}`,
		},
		{
			name: "values naming convention",
			data: `{"series":[{"columns":["time","rx"],"values":[["t1",1],["t2",2]]}]}`,
			want: `{"series":[{"columns":["time","rx"],"values":[["t2",2],["t1",1]]}]}`,
		},
		{
			name: "each series reversed independently",
			data: `{"series":[{"type":"a","data":[[1],[2]]},{"type":"b","data":[[3],[4]]}]}`,
			want: `{"series":[{"type":"a","data":[[2],[1]]},{"type":"b","data":[[4],[3]]}]}`,
		},
		{
			name: "single sample",
			data: `{"series":[{"data":[["t1",-100]]}]}`,
			want: `{"series":[{"data":[["t1",-100]]}]}`,
		},
		{
			name: "empty series list",
			data: `{"series":[]}`,
			want: `{"series":[]}`,
		},
		{
			name: "not series-shaped — unchanged",
			data: `{"result":[{"t":1},{"t":2}]}`,
			want: `{"result":[{"t":1},{"t":2}]}`,
		},
		{
			name: "bare array — unchanged",
			data: `[{"t":1},{"t":2}]`,
			want: `[{"t":1},{"t":2}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReverseSeriesData([]byte(tt.data))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !jsonEqual(t, got, []byte(tt.want)) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestFormatOutputStructuredTransformReachesMachineCallers is the regression for
// IM-3194: a reordering asked for by a flag has to reach json / yaml / --jq
// callers, not only the table renderer.
func TestFormatOutputStructuredTransformReachesMachineCallers(t *testing.T) {
	body := []byte(`{"result":{"series":[{"fields":["time"],"data":[["t1"],["t2"],["t3"]]}]}}`)

	t.Run("json output is reordered", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "json",
			WithTransform(FlattenSeries),
			WithStructuredTransform(ReverseSeriesData),
		); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); strings.Index(got, "t3") > strings.Index(got, "t1") {
			t.Errorf("expected newest sample first, got %s", got)
		}
	})

	t.Run("yaml output is reordered", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "yaml",
			WithTransform(FlattenSeries),
			WithStructuredTransform(ReverseSeriesData),
		); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); strings.Index(got, "t3") > strings.Index(got, "t1") {
			t.Errorf("expected newest sample first, got %s", got)
		}
	})

	t.Run("jq sees the reordered payload", func(t *testing.T) {
		io, out := newBufferIO()
		io.JQExpr = ".series[0].data[0][0]"
		if err := FormatOutput(body, io, "json",
			WithTransform(FlattenSeries),
			WithStructuredTransform(ReverseSeriesData),
		); err != nil {
			t.Fatal(err)
		}
		if got := strings.TrimSpace(out.String()); !strings.Contains(got, "t3") {
			t.Errorf("expected jq to see the newest sample first, got %s", got)
		}
	})

	t.Run("without a structured transform the payload is untouched", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "json", WithTransform(FlattenSeries)); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if strings.Index(got, "t1") > strings.Index(got, "t3") {
			t.Errorf("expected the original order to survive, got %s", got)
		}
		if !strings.Contains(got, "series") {
			t.Errorf("expected the series shape to survive, got %s", got)
		}
	})
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
