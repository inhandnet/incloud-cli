package iostreams

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/muesli/termenv"
)

// newBufferIO returns a non-TTY IOStreams writing into a buffer.
func newBufferIO() (*IOStreams, *bytes.Buffer) {
	var out bytes.Buffer
	io := &IOStreams{
		In:       os.Stdin,
		Out:      &out,
		ErrOut:   &out,
		outIsTTY: false,
		termOut:  termenv.NewOutput(&out, termenv.WithProfile(termenv.Ascii)),
	}
	return io, &out
}

// decode is a helper that parses JSON into a generic value for comparison.
func decode(t *testing.T, b []byte) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("invalid JSON %s: %v", b, err)
	}
	return v
}

func assertJSONEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	if !reflect.DeepEqual(decode(t, got), decode(t, []byte(want))) {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// testLatencyJitter and testOfflineDuration mirror the declarations in
// internal/unitdecl. They are duplicated here on purpose: iostreams is the
// mechanism and must be testable without knowing which commands use it.
var testLatencyJitter = FieldRewrites{
	"latency": {To: "latencyMicroseconds", StatusKey: "latencyStatus", Timeout: true},
	"jitter":  {To: "jitterMicroseconds", StatusKey: "jitterStatus", Timeout: true},
}

var testOfflineDuration = FieldRewrites{
	"totalOfflineDuration": {To: "totalOfflineDurationSeconds"},
	"avgOfflineDuration":   {To: "avgOfflineDurationSeconds"},
	"maxOfflineDuration":   {To: "maxOfflineDurationSeconds"},
}

func TestApplyFieldRewrites(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		rules FieldRewrites
		want  string
	}{
		{
			name:  "renames latency and jitter",
			in:    `{"name":"wan1","latency":42069,"jitter":120}`,
			rules: testLatencyJitter,
			want:  `{"name":"wan1","latencyMicroseconds":42069,"jitterMicroseconds":120}`,
		},
		{
			name:  "normal value gets no status field",
			in:    `{"latency":42069}`,
			rules: testLatencyJitter,
			want:  `{"latencyMicroseconds":42069}`,
		},
		{
			name:  "timeout sentinel becomes null with status",
			in:    `{"latency":-1,"jitter":-1}`,
			rules: testLatencyJitter,
			want: `{"latencyMicroseconds":null,"latencyStatus":"timeout",` +
				`"jitterMicroseconds":null,"jitterStatus":"timeout"}`,
		},
		{
			// 2000000 us = 2s is a plausible real measurement on a degraded
			// link. There is no clamp ceiling on this link, so it must pass
			// through unannotated (IM-3180 correction, 2026-08-17).
			name:  "two seconds is a real measurement and is not annotated",
			in:    `{"latency":2000000}`,
			rules: testLatencyJitter,
			want:  `{"latencyMicroseconds":2000000}`,
		},
		{
			name:  "four seconds is a real measurement and is not annotated",
			in:    `{"latency":4000000,"jitter":4000000}`,
			rules: testLatencyJitter,
			want:  `{"latencyMicroseconds":4000000,"jitterMicroseconds":4000000}`,
		},
		{
			name:  "missing fields are left alone",
			in:    `{"name":"wan1","state":"disconnected"}`,
			rules: testLatencyJitter,
			want:  `{"name":"wan1","state":"disconnected"}`,
		},
		{
			name:  "null value is renamed but not annotated",
			in:    `{"latency":null}`,
			rules: testLatencyJitter,
			want:  `{"latencyMicroseconds":null}`,
		},
		{
			name:  "non-numeric value is renamed but not annotated",
			in:    `{"latency":"n/a"}`,
			rules: testLatencyJitter,
			want:  `{"latencyMicroseconds":"n/a"}`,
		},
		{
			name: "nested and array entries are rewritten at any depth",
			in: `{"result":{"wan":[{"name":"wan1","latency":-1},{"name":"wan2","latency":33563}],` +
				`"cellular":[{"latency":2000000}]}}`,
			rules: testLatencyJitter,
			want: `{"result":{"wan":[{"name":"wan1","latencyMicroseconds":null,"latencyStatus":"timeout"},` +
				`{"name":"wan2","latencyMicroseconds":33563}],"cellular":[{"latencyMicroseconds":2000000}]}}`,
		},
		{
			name:  "offline durations get a seconds suffix",
			in:    `{"result":[{"totalOfflineDuration":173114,"avgOfflineDuration":10,"maxOfflineDuration":924}]}`,
			rules: testOfflineDuration,
			want: `{"result":[{"totalOfflineDurationSeconds":173114,"avgOfflineDurationSeconds":10,` +
				`"maxOfflineDurationSeconds":924}]}`,
		},
		{
			name:  "offline durations get no sentinel annotation",
			in:    `{"maxOfflineDuration":-1}`,
			rules: testOfflineDuration,
			want:  `{"maxOfflineDurationSeconds":-1}`,
		},
		{
			name:  "empty rules leave the body untouched",
			in:    `{"latency":42069}`,
			rules: FieldRewrites{},
			want:  `{"latency":42069}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertJSONEqual(t, applyFieldRewrites([]byte(tc.in), tc.rules), tc.want)
		})
	}
}

func TestApplyFieldRewritesInvalidJSON(t *testing.T) {
	in := []byte("not json at all")
	got := applyFieldRewrites(in, testLatencyJitter)
	if string(got) != string(in) {
		t.Errorf("invalid JSON should pass through unchanged, got %s", got)
	}
}

func TestApplyFieldRewritesColumnarSeries(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "columns are renamed, no status column when all values are normal",
			in: `{"series":[{"name":"uplink","columns":["time","latency","jitter","loss"],` +
				`"values":[["t1",33685,120,0],["t2",43972,130,0]]}]}`,
			want: `{"series":[{"name":"uplink","columns":["time","latencyMicroseconds","jitterMicroseconds","loss"],` +
				`"values":[["t1",33685,120,0],["t2",43972,130,0]]}]}`,
		},
		{
			name: "sentinels append a status column aligned with the rows",
			in: `{"series":[{"columns":["time","latency"],` +
				`"values":[["t1",-1],["t2",42069],["t3",2000000]]}]}`,
			want: `{"series":[{"columns":["time","latencyMicroseconds","latencyStatus"],` +
				`"values":[["t1",null,"timeout"],["t2",42069,null],["t3",2000000,null]]}]}`,
		},
		{
			name: "latency and jitter sentinels append two aligned status columns",
			in: `{"series":[{"columns":["time","latency","jitter"],` +
				`"values":[["t1",-1,-1],["t2",42069,120]]}]}`,
			want: `{"series":[{"columns":["time","latencyMicroseconds","jitterMicroseconds",` +
				`"latencyStatus","jitterStatus"],` +
				`"values":[["t1",null,null,"timeout","timeout"],["t2",42069,120,null,null]]}]}`,
		},
		{
			name: "fields/data naming convention is supported too",
			in:   `{"series":[{"fields":["time","latency"],"data":[["t1",-1]]}]}`,
			want: `{"series":[{"fields":["time","latencyMicroseconds","latencyStatus"],"data":[["t1",null,"timeout"]]}]}`,
		},
		{
			name: "fields/data without sentinels only renames the field names",
			in:   `{"series":[{"fields":["time","latency","jitter"],"data":[["t1",33685,120]]}]}`,
			want: `{"series":[{"fields":["time","latencyMicroseconds","jitterMicroseconds"],"data":[["t1",33685,120]]}]}`,
		},
		{
			name: "null samples pass through without an annotation",
			in:   `{"series":[{"columns":["time","latency","jitter"],"values":[["t1",33685,null]]}]}`,
			want: `{"series":[{"columns":["time","latencyMicroseconds","jitterMicroseconds"],"values":[["t1",33685,null]]}]}`,
		},
		{
			name: "empty value set still renames the columns",
			in:   `{"series":[{"columns":["time","latency"],"values":[]}]}`,
			want: `{"series":[{"columns":["time","latencyMicroseconds"],"values":[]}]}`,
		},
		{
			name: "empty column list is left alone",
			in:   `{"series":[{"columns":[],"values":[["t1",42069]]}]}`,
			want: `{"series":[{"columns":[],"values":[["t1",42069]]}]}`,
		},
		{
			name: "non-string column names are skipped",
			in:   `{"series":[{"columns":[7,"latency"],"values":[["t1",42069]]}]}`,
			want: `{"series":[{"columns":[7,"latencyMicroseconds"],"values":[["t1",42069]]}]}`,
		},
		{
			name: "series whose columns match no rule is untouched",
			in:   `{"series":[{"columns":["time","signal"],"values":[["t1",-1]]}]}`,
			want: `{"series":[{"columns":["time","signal"],"values":[["t1",-1]]}]}`,
		},
		{
			name: "rows that are not arrays are left in place",
			in:   `{"series":[{"columns":["time","latency"],"values":[["t1",-1],"broken"]}]}`,
			want: `{"series":[{"columns":["time","latencyMicroseconds","latencyStatus"],` +
				`"values":[["t1",null,"timeout"],"broken"]}]}`,
		},
		{
			name: "columns without a matching values array is not a series",
			in:   `{"series":[{"columns":["time","latency"],"values":{"t1":-1}}]}`,
			want: `{"series":[{"columns":["time","latency"],"values":{"t1":-1}}]}`,
		},
		{
			name: "duplicate column names are both rewritten",
			in:   `{"series":[{"columns":["latency","latency"],"values":[[-1,42069]]}]}`,
			want: `{"series":[{"columns":["latencyMicroseconds","latencyMicroseconds","latencyStatus"],` +
				`"values":[[null,42069,"timeout"]]}]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertJSONEqual(t, applyFieldRewrites([]byte(tc.in), testLatencyJitter), tc.want)
		})
	}
}

// TestColumnarShortRowsStayAligned covers the column-misalignment bug: a row
// shorter than "columns" used to be skipped while building values but still got
// a status appended, landing the status in a value column.
func TestColumnarShortRowsStayAligned(t *testing.T) {
	cases := []string{
		// short row, some other row carries a sentinel
		`{"columns":["time","latency"],"values":[["t0",-1],["t1"]]}`,
		// short row, no sentinel anywhere
		`{"columns":["time","latency"],"values":[["t0",2000000],["t1"]]}`,
		// short row, latency and jitter both sentinel elsewhere
		`{"columns":["time","latency","jitter"],"values":[["t0",-1,-1],["t1"],["t2",1,2]]}`,
	}
	for _, in := range cases {
		got := applyFieldRewrites([]byte(in), testLatencyJitter)
		var out struct {
			Columns []interface{}   `json:"columns"`
			Values  [][]interface{} `json:"values"`
		}
		if err := json.Unmarshal(got, &out); err != nil {
			t.Fatalf("%s -> %s: %v", in, got, err)
		}
		for i, row := range out.Values {
			if len(row) != len(out.Columns) {
				t.Errorf("%s -> %s: row %d has %d cells, want %d",
					in, got, i, len(row), len(out.Columns))
			}
		}
	}
}

func TestAsFloat(t *testing.T) {
	tests := []struct {
		name string
		raw  interface{}
		want float64
		ok   bool
	}{
		{name: "json.Number integer", raw: json.Number("42069"), want: 42069, ok: true},
		{name: "json.Number unparsable", raw: json.Number("not-a-number"), ok: false},
		{name: "float64 passes through", raw: float64(-1), want: -1, ok: true},
		{name: "string is not numeric", raw: "-1", ok: false},
		{name: "nil is not numeric", raw: nil, ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := asFloat(tc.raw)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Errorf("asFloat(%v) = (%v, %v), want (%v, %v)", tc.raw, got, ok, tc.want, tc.ok)
			}
		})
	}
}

// TestApplyRuleAcceptsFloat64 covers the float64 branch of asFloat, which is
// reached when the tree was decoded without UseNumber (e.g. by a caller that
// pre-parsed the body).
func TestApplyRuleAcceptsFloat64(t *testing.T) {
	value, status := applyRule(float64(-1), testLatencyJitter["latency"])
	if value != nil || status != StatusTimeout {
		t.Errorf("got (%v, %q), want (nil, %q)", value, status, StatusTimeout)
	}
	value, status = applyRule(json.Number("bogus"), testLatencyJitter["latency"])
	if status != "" {
		t.Errorf("unparsable number should not be annotated, got %q", status)
	}
	if value != json.Number("bogus") {
		t.Errorf("unparsable number should pass through, got %v", value)
	}
}

// TestFormatOutputWithoutDeclarationIsUnchanged covers acceptance criterion 9:
// without a declaration, structured output carries the input field names and
// values through unchanged, sentinel included.
func TestFormatOutputWithoutDeclarationIsUnchanged(t *testing.T) {
	body := []byte(`{"result":[{"name":"wan1","latency":-1}]}`)
	io, out := newBufferIO()
	if err := FormatOutput(body, io, "json"); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, `"latency"`) || !strings.Contains(got, "-1") {
		t.Errorf("expected the input field and value to survive, got %s", got)
	}
	for _, absent := range []string{"latencyMicroseconds", "latencyStatus"} {
		if strings.Contains(got, absent) {
			t.Errorf("expected no %s without a declaration, got %s", absent, got)
		}
	}
}

func TestFormatOutputRewritesStructuredButNotTable(t *testing.T) {
	body := []byte(`{"result":[{"name":"wan1","latency":42069}]}`)

	t.Run("json output is rewritten", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "json",
			WithFormatters(ColumnFormatters{"latency": FormatMicroseconds}),
			WithDeclaredUnits(testLatencyJitter),
		); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, "latencyMicroseconds") || strings.Contains(got, `"latency"`) {
			t.Errorf("expected latencyMicroseconds and no bare latency, got %s", got)
		}
	})

	t.Run("yaml output is rewritten", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "yaml",
			WithDeclaredUnits(testLatencyJitter),
		); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, "latencyMicroseconds") {
			t.Errorf("expected latencyMicroseconds in yaml, got %s", got)
		}
	})

	t.Run("jq sees the rewritten field", func(t *testing.T) {
		io, out := newBufferIO()
		io.JQExpr = ".[0].latencyMicroseconds"
		if err := FormatOutput(body, io, "json",
			WithDeclaredUnits(testLatencyJitter),
		); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "42069") {
			t.Errorf("expected jq to resolve latencyMicroseconds, got %s", out.String())
		}
	})

	t.Run("jq no longer sees the old field name", func(t *testing.T) {
		io, out := newBufferIO()
		io.JQExpr = ".[0].latency"
		if err := FormatOutput(body, io, "json",
			WithDeclaredUnits(testLatencyJitter),
		); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out.String(), "42069") {
			t.Errorf("expected .latency to be absent, got %s", out.String())
		}
	})

	t.Run("table output is byte-identical with and without rewrites", func(t *testing.T) {
		ioA, outA := newBufferIO()
		if err := FormatOutput(body, ioA, "table",
			WithFormatters(ColumnFormatters{"latency": FormatMicroseconds}),
		); err != nil {
			t.Fatal(err)
		}
		ioB, outB := newBufferIO()
		if err := FormatOutput(body, ioB, "table",
			WithFormatters(ColumnFormatters{"latency": FormatMicroseconds}),
			WithDeclaredUnits(testLatencyJitter),
		); err != nil {
			t.Fatal(err)
		}
		if outA.String() != outB.String() {
			t.Errorf("table output changed:\nwithout: %q\nwith:    %q", outA.String(), outB.String())
		}
		if !strings.Contains(outA.String(), "42.069 ms") {
			t.Errorf("expected table to still render 42.069 ms, got %q", outA.String())
		}
	})
}

// TestApplyFieldRewritesPreservesIntegerPrecision guards against the rewriter
// round-tripping numbers through float64, which would corrupt large integers.
func TestApplyFieldRewritesPreservesIntegerPrecision(t *testing.T) {
	got := applyFieldRewrites([]byte(`{"latency":42069,"_id":9007199254740993}`), testLatencyJitter)
	for _, want := range []string{`"latencyMicroseconds":42069`, `9007199254740993`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("expected %s in %s", want, got)
		}
	}
}
