package iostreams

import (
	"strings"
	"testing"
)

// --- FormatBytes tests ---

func TestFormatBytes_Zero(t *testing.T) {
	if got := FormatBytes("0"); got != "0 B" {
		t.Errorf("expected '0 B', got %q", got)
	}
}

func TestFormatBytes_Bytes(t *testing.T) {
	if got := FormatBytes("512"); got != "512 B" {
		t.Errorf("expected '512 B', got %q", got)
	}
}

func TestFormatBytes_KiB(t *testing.T) {
	if got := FormatBytes("1024"); got != "1.0 KiB" {
		t.Errorf("expected '1.0 KiB', got %q", got)
	}
}

func TestFormatBytes_MiB(t *testing.T) {
	if got := FormatBytes("1048576"); got != "1.0 MiB" {
		t.Errorf("expected '1.0 MiB', got %q", got)
	}
}

func TestFormatBytes_GiB(t *testing.T) {
	if got := FormatBytes("21114126336"); got != "20 GiB" {
		t.Errorf("expected '20 GiB', got %q", got)
	}
}

func TestFormatBytes_TiB(t *testing.T) {
	if got := FormatBytes("1099511627776"); got != "1.0 TiB" {
		t.Errorf("expected '1.0 TiB', got %q", got)
	}
}

func TestFormatBytes_NotANumber(t *testing.T) {
	if got := FormatBytes("N/A"); got != "N/A" {
		t.Errorf("expected 'N/A', got %q", got)
	}
}

func TestFormatBytes_Empty(t *testing.T) {
	if got := FormatBytes(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// --- FormatPercent tests ---

func TestFormatPercent_Half(t *testing.T) {
	if got := FormatPercent("0.5"); got != "50.0%" {
		t.Errorf("expected '50.0%%', got %q", got)
	}
}

func TestFormatPercent_Full(t *testing.T) {
	if got := FormatPercent("1.0"); got != "100.0%" {
		t.Errorf("expected '100.0%%', got %q", got)
	}
}

func TestFormatPercent_Small(t *testing.T) {
	if got := FormatPercent("0.452"); got != "45.2%" {
		t.Errorf("expected '45.2%%', got %q", got)
	}
}

func TestFormatPercent_NotANumber(t *testing.T) {
	if got := FormatPercent("err"); got != "err" {
		t.Errorf("expected 'err', got %q", got)
	}
}

// --- FormatDuration tests ---

func TestFormatDuration_Seconds(t *testing.T) {
	if got := FormatDuration("45"); got != "45s" {
		t.Errorf("expected '45s', got %q", got)
	}
}

func TestFormatDuration_Minutes(t *testing.T) {
	if got := FormatDuration("90"); got != "1m 30s" {
		t.Errorf("expected '1m 30s', got %q", got)
	}
}

func TestFormatDuration_Hours(t *testing.T) {
	if got := FormatDuration("3661"); got != "1h 1m" {
		t.Errorf("expected '1h 1m', got %q", got)
	}
}

func TestFormatDuration_Days(t *testing.T) {
	if got := FormatDuration("90061"); got != "1d 1h 1m" {
		t.Errorf("expected '1d 1h 1m', got %q", got)
	}
}

func TestFormatDuration_Milliseconds(t *testing.T) {
	if got := FormatDuration("0.5"); got != "500ms" {
		t.Errorf("expected '500ms', got %q", got)
	}
}

func TestFormatDuration_Zero(t *testing.T) {
	if got := FormatDuration("0"); got != "0s" {
		t.Errorf("expected '0s', got %q", got)
	}
}

func TestFormatDuration_NotANumber(t *testing.T) {
	if got := FormatDuration("abc"); got != "abc" {
		t.Errorf("expected 'abc', got %q", got)
	}
}

// --- FormatBitRate tests ---

func TestFormatBitRate_Bps(t *testing.T) {
	if got := FormatBitRate("500"); got != "500 bps" {
		t.Errorf("expected '500 bps', got %q", got)
	}
}

func TestFormatBitRate_Kbps(t *testing.T) {
	if got := FormatBitRate("1500"); got != "1.5 kbps" {
		t.Errorf("expected '1.5 kbps', got %q", got)
	}
}

func TestFormatBitRate_Mbps(t *testing.T) {
	if got := FormatBitRate("1000000"); got != "1 Mbps" {
		t.Errorf("expected '1 Mbps', got %q", got)
	}
}

func TestFormatBitRate_Gbps(t *testing.T) {
	if got := FormatBitRate("2500000000"); got != "2.5 Gbps" {
		t.Errorf("expected '2.5 Gbps', got %q", got)
	}
}

func TestFormatBitRate_NotANumber(t *testing.T) {
	if got := FormatBitRate("N/A"); got != "N/A" {
		t.Errorf("expected 'N/A', got %q", got)
	}
}

// --- FormatMbps tests ---

func TestFormatMbps_Normal(t *testing.T) {
	if got := FormatMbps("25.5"); got != "25.50 Mbps" {
		t.Errorf("expected '25.50 Mbps', got %q", got)
	}
}

// --- FormatMicroseconds tests ---

func TestFormatMicroseconds_Normal(t *testing.T) {
	if got := FormatMicroseconds("4765"); got != "4.765 ms" {
		t.Errorf("expected '4.765 ms', got %q", got)
	}
}

func TestFormatMicroseconds_Small(t *testing.T) {
	if got := FormatMicroseconds("500"); got != "0.500 ms" {
		t.Errorf("expected '0.500 ms', got %q", got)
	}
}

func TestFormatMicroseconds_NotANumber(t *testing.T) {
	if got := FormatMicroseconds("N/A"); got != "N/A" {
		t.Errorf("expected 'N/A', got %q", got)
	}
}

// --- FormatMs tests ---

func TestFormatMs_Normal(t *testing.T) {
	if got := FormatMs("12.5"); got != "12.50 ms" {
		t.Errorf("expected '12.50 ms', got %q", got)
	}
}

func TestFormatMs_Zero(t *testing.T) {
	if got := FormatMs("0"); got != "0.00 ms" {
		t.Errorf("expected '0.00 ms', got %q", got)
	}
}

func TestFormatMs_NotANumber(t *testing.T) {
	if got := FormatMs("err"); got != "err" {
		t.Errorf("expected 'err', got %q", got)
	}
}

func TestFormatMbps_Zero(t *testing.T) {
	if got := FormatMbps("0"); got != "0.00 Mbps" {
		t.Errorf("expected '0.00 Mbps', got %q", got)
	}
}

func TestFormatMbps_NotANumber(t *testing.T) {
	if got := FormatMbps("err"); got != "err" {
		t.Errorf("expected 'err', got %q", got)
	}
}

// --- WithFormatters integration tests ---

func TestFormatOutput_WithFormatters_Array(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","tx":1048576,"rx":2097152},{"name":"dev2","tx":512,"rx":1024}]}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"tx": FormatBytes,
		"rx": FormatBytes,
	}
	err := FormatOutput(data, io, "table", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.0 MiB") {
		t.Errorf("expected tx=1.0 MiB, got:\n%s", out)
	}
	if !strings.Contains(out, "2.0 MiB") {
		t.Errorf("expected rx=2.0 MiB, got:\n%s", out)
	}
	if !strings.Contains(out, "512 B") {
		t.Errorf("expected tx=512 B, got:\n%s", out)
	}
	if !strings.Contains(out, "1.0 KiB") {
		t.Errorf("expected rx=1.0 KiB, got:\n%s", out)
	}
	// name should not be formatted
	if !strings.Contains(out, "dev1") {
		t.Errorf("expected dev1 unchanged, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_Object(t *testing.T) {
	data := []byte(`{"result":{"name":"dev1","tx":1048576}}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"tx": FormatBytes,
	}
	err := FormatOutput(data, io, "table", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.0 MiB") {
		t.Errorf("expected tx=1.0 MiB, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_JSON_NotApplied(t *testing.T) {
	data := []byte(`{"result":[{"tx":1048576}]}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"tx": FormatBytes,
	}
	err := FormatOutput(data, io, "json", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// JSON output should NOT apply formatters
	if strings.Contains(out, "MiB") {
		t.Errorf("json output should not apply formatters, got:\n%s", out)
	}
	if !strings.Contains(out, "1048576") {
		t.Errorf("expected raw value in json output, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_AndTransform(t *testing.T) {
	// Simulate a series response that gets flattened, then formatted
	data := []byte(`{"result":{"series":[{"fields":["time","tx","rx"],"data":[["2026-01-01",1048576,2097152]]}]}}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"tx": FormatBytes,
		"rx": FormatBytes,
	}
	err := FormatOutput(data, io, "table", WithTransform(FlattenSeries), WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.0 MiB") {
		t.Errorf("expected tx=1.0 MiB after transform+format, got:\n%s", out)
	}
	if !strings.Contains(out, "2.0 MiB") {
		t.Errorf("expected rx=2.0 MiB after transform+format, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_Percent(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","cpu.usage":0.452}]}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"cpu.usage": FormatPercent,
	}
	err := FormatOutput(data, io, "table", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "45.2%") {
		t.Errorf("expected 45.2%%, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_NestedDotPath(t *testing.T) {
	data := []byte(`{"result":[{"deviceId":"d1","sim":{"tx":1048576,"rx":2097152,"total":3145728}}]}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"sim.tx":    FormatBytes,
		"sim.rx":    FormatBytes,
		"sim.total": FormatBytes,
	}
	err := FormatOutput(data, io, "table", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "1.0 MiB") {
		t.Errorf("expected sim.tx=1.0 MiB, got:\n%s", out)
	}
	if !strings.Contains(out, "2.0 MiB") {
		t.Errorf("expected sim.rx=2.0 MiB, got:\n%s", out)
	}
	if !strings.Contains(out, "3.0 MiB") {
		t.Errorf("expected sim.total=3.0 MiB, got:\n%s", out)
	}
}

func TestFormatOutput_WithFormatters_NoMatch(t *testing.T) {
	data := []byte(`{"result":[{"name":"dev1","tx":1024}]}`)
	io, buf := newTestIOWithBuf(false)
	fmts := ColumnFormatters{
		"nonexistent": FormatBytes,
	}
	err := FormatOutput(data, io, "table", WithFormatters(fmts))
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// tx should remain unformatted
	if !strings.Contains(out, "1024") {
		t.Errorf("expected raw 1024, got:\n%s", out)
	}
}
