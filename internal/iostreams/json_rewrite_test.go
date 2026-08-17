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
			rules: LatencyJitterRewrites,
			want:  `{"name":"wan1","latencyUs":42069,"jitterUs":120}`,
		},
		{
			name:  "normal value gets no status field",
			in:    `{"latency":42069}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":42069}`,
		},
		{
			name:  "no-measurement sentinel becomes null with status",
			in:    `{"latency":-1,"jitter":-1}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":null,"latencyStatus":"no-measurement","jitterUs":null,"jitterStatus":"no-measurement"}`,
		},
		{
			name:  "2s clamp keeps original value and is annotated",
			in:    `{"latency":2000000}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":2000000,"latencyStatus":"suspected-timeout"}`,
		},
		{
			name:  "4s clamp keeps original value and is annotated",
			in:    `{"latency":4000000}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":4000000,"latencyStatus":"suspected-timeout"}`,
		},
		{
			name:  "jitter does not get the timeout clamp treatment",
			in:    `{"jitter":2000000}`,
			rules: LatencyJitterRewrites,
			want:  `{"jitterUs":2000000}`,
		},
		{
			name:  "missing fields are left alone",
			in:    `{"name":"wan1","state":"disconnected"}`,
			rules: LatencyJitterRewrites,
			want:  `{"name":"wan1","state":"disconnected"}`,
		},
		{
			name:  "null value is renamed but not annotated",
			in:    `{"latency":null}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":null}`,
		},
		{
			name:  "non-numeric value is renamed but not annotated",
			in:    `{"latency":"n/a"}`,
			rules: LatencyJitterRewrites,
			want:  `{"latencyUs":"n/a"}`,
		},
		{
			name:  "nested and array entries are rewritten at any depth",
			in:    `{"result":{"wan":[{"name":"wan1","latency":-1},{"name":"wan2","latency":33563}],"cellular":[{"latency":2000000}]}}`,
			rules: LatencyJitterRewrites,
			want: `{"result":{"wan":[{"name":"wan1","latencyUs":null,"latencyStatus":"no-measurement"},` +
				`{"name":"wan2","latencyUs":33563}],"cellular":[{"latencyUs":2000000,"latencyStatus":"suspected-timeout"}]}}`,
		},
		{
			name:  "offline durations get a seconds suffix",
			in:    `{"result":[{"totalOfflineDuration":173114,"avgOfflineDuration":10,"maxOfflineDuration":924}]}`,
			rules: OfflineDurationRewrites,
			want:  `{"result":[{"totalOfflineDurationSeconds":173114,"avgOfflineDurationSeconds":10,"maxOfflineDurationSeconds":924}]}`,
		},
		{
			name:  "offline durations get no sentinel annotation",
			in:    `{"maxOfflineDuration":-1}`,
			rules: OfflineDurationRewrites,
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
	got := applyFieldRewrites(in, LatencyJitterRewrites)
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
			want: `{"series":[{"name":"uplink","columns":["time","latencyUs","jitterUs","loss"],` +
				`"values":[["t1",33685,120,0],["t2",43972,130,0]]}]}`,
		},
		{
			name: "sentinels append a status column aligned with the rows",
			in: `{"series":[{"columns":["time","latency"],` +
				`"values":[["t1",-1],["t2",42069],["t3",2000000]]}]}`,
			want: `{"series":[{"columns":["time","latencyUs","latencyStatus"],` +
				`"values":[["t1",null,"no-measurement"],["t2",42069,null],["t3",2000000,"suspected-timeout"]]}]}`,
		},
		{
			name: "fields/data naming convention is supported too",
			in:   `{"series":[{"fields":["time","latency"],"data":[["t1",-1]]}]}`,
			want: `{"series":[{"fields":["time","latencyUs","latencyStatus"],"data":[["t1",null,"no-measurement"]]}]}`,
		},
		{
			name: "null samples pass through without an annotation",
			in:   `{"series":[{"columns":["time","latency","jitter"],"values":[["t1",33685,null]]}]}`,
			want: `{"series":[{"columns":["time","latencyUs","jitterUs"],"values":[["t1",33685,null]]}]}`,
		},
		{
			name: "empty value set still renames the columns",
			in:   `{"series":[{"columns":["time","latency"],"values":[]}]}`,
			want: `{"series":[{"columns":["time","latencyUs"],"values":[]}]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertJSONEqual(t, applyFieldRewrites([]byte(tc.in), LatencyJitterRewrites), tc.want)
		})
	}
}

func TestFormatOutputRewritesStructuredButNotTable(t *testing.T) {
	body := []byte(`{"result":[{"name":"wan1","latency":42069}]}`)

	t.Run("json output is rewritten", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "json",
			WithFormatters(ColumnFormatters{"latency": FormatMicroseconds}),
			WithJSONFieldRewrites(LatencyJitterRewrites),
		); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, "latencyUs") || strings.Contains(got, `"latency"`) {
			t.Errorf("expected latencyUs and no bare latency, got %s", got)
		}
	})

	t.Run("yaml output is rewritten", func(t *testing.T) {
		io, out := newBufferIO()
		if err := FormatOutput(body, io, "yaml",
			WithJSONFieldRewrites(LatencyJitterRewrites),
		); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, "latencyUs") {
			t.Errorf("expected latencyUs in yaml, got %s", got)
		}
	})

	t.Run("jq sees the rewritten field", func(t *testing.T) {
		io, out := newBufferIO()
		io.JQExpr = ".[0].latencyUs"
		if err := FormatOutput(body, io, "json",
			WithJSONFieldRewrites(LatencyJitterRewrites),
		); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "42069") {
			t.Errorf("expected jq to resolve latencyUs, got %s", out.String())
		}
	})

	t.Run("jq no longer sees the old field name", func(t *testing.T) {
		io, out := newBufferIO()
		io.JQExpr = ".[0].latency"
		if err := FormatOutput(body, io, "json",
			WithJSONFieldRewrites(LatencyJitterRewrites),
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
			WithJSONFieldRewrites(LatencyJitterRewrites),
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
	got := applyFieldRewrites([]byte(`{"latency":42069,"_id":9007199254740993}`), LatencyJitterRewrites)
	for _, want := range []string{`"latencyUs":42069`, `9007199254740993`} {
		if !strings.Contains(string(got), want) {
			t.Errorf("expected %s in %s", want, got)
		}
	}
}
