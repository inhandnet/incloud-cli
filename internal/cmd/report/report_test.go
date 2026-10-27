package report

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

const testID = "507f1f77bcf86cd799439011"

func newTestFactory(t *testing.T, host string) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {Host: host, Token: "test-token"},
		},
	}
	if _, err := config.Update(cfgPath, func(c *config.Config) error { *c = *cfg; return nil }); err != nil {
		t.Fatal(err)
	}

	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	f := &factory.Factory{
		IO:         &iostreams.IOStreams{In: strings.NewReader(""), Out: out, ErrOut: errBuf},
		ConfigPath: cfgPath,
	}
	return f, out, errBuf
}

func newRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdReport(f))
	return root
}

type captured struct {
	method, path, query string
	body                map[string]any
}

// recorder returns a server that records requests and answers with resp.
func recorder(t *testing.T, resp any) (*httptest.Server, *[]captured) {
	t.Helper()
	var reqs []captured
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := captured{method: r.Method, path: r.URL.Path, query: r.URL.RawQuery}
		if data, _ := io.ReadAll(r.Body); len(data) > 0 {
			if err := json.Unmarshal(data, &c.body); err != nil {
				t.Errorf("request body: %v", err)
			}
		}
		reqs = append(reqs, c)
		json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv, &reqs
}

func run(t *testing.T, host string, args ...string) (string, string, error) {
	t.Helper()
	f, out, errBuf := newTestFactory(t, host)
	root := newRoot(f)
	root.SetArgs(append([]string{"report"}, args...))
	err := root.Execute()
	return out.String(), errBuf.String(), err
}

func TestPolicyCreate_Bodies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "daily",
			args: []string{"--name", "d", "--template", "total_devices", "--period", "last_1_day",
				"--frequency", "daily", "--at", "09:05", "--timezone", "Asia/Tokyo", "--recipient", "a@b.c"},
			want: map[string]any{
				"name": "d", "applyTo": map[string]any{"type": "ALL"}, "templates": []any{"TOTAL_DEVICES"},
				"format": "PDF", "period": "LAST_1_DAY", "scheduled": true,
				"schedule":  map[string]any{"hourOfDay": float64(9), "minuteOfHour": float64(5)},
				"timezone":  "Asia/Tokyo",
				"recipient": "a@b.c",
			},
		},
		{
			name: "weekly by name, group targets",
			args: []string{"--name", "w", "--apply-to", "group", "--target", testID,
				"--template", "DEVICE_DATA_USAGE_DETAILS", "--format", "xlsx", "--period", "last_7_days",
				"--frequency", "weekly", "--day-of-week", "sunday", "--at", "23:59"},
			want: map[string]any{
				"name": "w", "applyTo": map[string]any{"type": "GROUP", "ids": []any{testID}},
				"templates": []any{"DEVICE_DATA_USAGE_DETAILS"}, "format": "XLSX", "period": "LAST_7_DAYS",
				"scheduled": true,
				"schedule":  map[string]any{"hourOfDay": float64(23), "minuteOfHour": float64(59), "dayOfWeek": float64(7)},
			},
		},
		{
			name: "monthly",
			args: []string{"--name", "m", "--template", "USAGE_SUMMARY", "--period", "last_30_days",
				"--frequency", "monthly", "--day-of-month", "28", "--at", "00:00"},
			want: map[string]any{
				"name": "m", "applyTo": map[string]any{"type": "ALL"}, "templates": []any{"USAGE_SUMMARY"},
				"format": "PDF", "period": "LAST_30_DAYS", "scheduled": true,
				"schedule": map[string]any{"hourOfDay": float64(0), "minuteOfHour": float64(0), "dayOfMonth": float64(28)},
			},
		},
		{
			name: "one-off range in timezone",
			args: []string{"--name", "o", "--template", "USAGE_SUMMARY",
				"--from", "2026-10-01", "--to", "2026-10-31", "--timezone", "Asia/Tokyo"},
			want: map[string]any{
				"name": "o", "applyTo": map[string]any{"type": "ALL"}, "templates": []any{"USAGE_SUMMARY"},
				"format": "PDF", "scheduled": false, "timezone": "Asia/Tokyo",
				"startTime": "2026-09-30T15:00:00Z", "endTime": "2026-10-31T15:00:00Z",
			},
		},
		{
			name: "one-off period",
			args: []string{"--name", "p", "--template", "USAGE_SUMMARY", "--period", "last_3_days"},
			want: map[string]any{
				"name": "p", "applyTo": map[string]any{"type": "ALL"}, "templates": []any{"USAGE_SUMMARY"},
				"format": "PDF", "scheduled": false, "period": "LAST_3_DAYS",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, reqs := recorder(t, map[string]any{"result": map[string]any{"_id": testID}})
			_, stderr, err := run(t, srv.URL, append([]string{"policy", "create", "-o", "json"}, tt.args...)...)
			if err != nil {
				t.Fatal(err)
			}
			if len(*reqs) != 1 || (*reqs)[0].method != http.MethodPost || (*reqs)[0].path != policiesPath {
				t.Fatalf("requests = %+v", *reqs)
			}
			if got := (*reqs)[0].body; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("body =\n%v\nwant\n%v", got, tt.want)
			}
			if !strings.Contains(stderr, testID) {
				t.Errorf("stderr %q should mention the new id", stderr)
			}
		})
	}
}

func TestPolicyCreate_LocalErrors(t *testing.T) {
	base := []string{"--name", "x", "--template", "USAGE_SUMMARY"}
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"bad timezone", append(base, "--period", "last_1_day", "--timezone", "Mars/Base"), "invalid --timezone"},
		{"local timezone", append(base, "--period", "last_1_day", "--timezone", "Local"), "invalid --timezone"},
		{"unknown template", []string{"--name", "x", "--template", "NOPE", "--period", "last_1_day"}, "unknown --template"},
		{"xlsx only with pdf", []string{"--name", "x", "--template", "DEVICE_ONLINE_RATE", "--period", "last_1_day"}, "only available in xlsx"},
		{"no name", []string{"--template", "USAGE_SUMMARY", "--period", "last_1_day"}, "--name is required"},
		{"no template", []string{"--name", "x", "--period", "last_1_day"}, "--template is required"},
		{"group without target", append(base, "--apply-to", "group", "--period", "last_1_day"), "requires at least one --target"},
		{"target with all", append(base, "--target", testID, "--period", "last_1_day"), "--target is only used"},
		{"bad target", append(base, "--apply-to", "org", "--target", "abc", "--period", "last_1_day"), "invalid --target"},
		{"scheduled without period", append(base, "--frequency", "daily", "--at", "08:00"), "require --period"},
		{"scheduled without at", append(base, "--frequency", "daily", "--period", "last_1_day"), "requires --at"},
		{"bad at", append(base, "--frequency", "daily", "--at", "25:00", "--period", "last_1_day"), "invalid --at"},
		{"weekly without day", append(base, "--frequency", "weekly", "--at", "08:00", "--period", "last_7_days"), "requires --day-of-week"},
		{"bad day of week", append(base, "--frequency", "weekly", "--day-of-week", "8", "--at", "08:00", "--period", "last_7_days"), "invalid --day-of-week"},
		{"day of month range", append(base, "--frequency", "monthly", "--day-of-month", "31", "--at", "08:00", "--period", "last_30_days"), "between 1 and 28"},
		{"day of week on daily", append(base, "--frequency", "daily", "--day-of-week", "1", "--at", "08:00", "--period", "last_1_day"), "only used with --frequency weekly"},
		{"at without frequency", append(base, "--at", "08:00", "--period", "last_1_day"), "require --frequency"},
		{"from with frequency", append(base, "--frequency", "daily", "--at", "08:00", "--period", "last_1_day", "--from", "2026-10-01", "--to", "2026-10-02"), "cannot be combined"},
		{"from only", append(base, "--from", "2026-10-01"), "both --from and --to"},
		{"to before from", append(base, "--from", "2026-10-02", "--to", "2026-10-01"), "must not be before"},
		{"nothing", base, "pass --frequency"},
		{"period with range", append(base, "--period", "last_1_day", "--from", "2026-10-01", "--to", "2026-10-02"), "cannot be combined"},
		{"bad period", append(base, "--period", "last_2_days"), "invalid --period"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, reqs := recorder(t, map[string]any{})
			_, _, err := run(t, srv.URL, append([]string{"policy", "create"}, tt.args...)...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
			if len(*reqs) != 0 {
				t.Errorf("no request expected, got %+v", *reqs)
			}
		})
	}
}

func TestPolicyUpdate_SendsOnlyChangedFields(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{"timezone only", []string{"--timezone", "America/New_York"}, map[string]any{"timezone": "America/New_York"}},
		{"schedule", []string{"--frequency", "weekly", "--day-of-week", "2", "--at", "07:30"},
			map[string]any{"schedule": map[string]any{"hourOfDay": float64(7), "minuteOfHour": float64(30), "dayOfWeek": float64(2)}}},
		{"apply-to all", []string{"--apply-to", "all", "--recipient", ""},
			map[string]any{"applyTo": map[string]any{"type": "ALL"}, "recipient": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, reqs := recorder(t, map[string]any{"result": map[string]any{"_id": testID}})
			if _, _, err := run(t, srv.URL, append([]string{"policy", "update", testID, "-o", "json"}, tt.args...)...); err != nil {
				t.Fatal(err)
			}
			if len(*reqs) != 1 || (*reqs)[0].method != http.MethodPut || (*reqs)[0].path != policiesPath+"/"+testID {
				t.Fatalf("requests = %+v", *reqs)
			}
			if got := (*reqs)[0].body; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("body = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolicyUpdate_LocalErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"clear timezone", []string{"--timezone", ""}, "cannot be cleared"},
		{"bad timezone", []string{"--timezone", "Nowhere"}, "invalid --timezone"},
		{"partial schedule", []string{"--at", "08:00"}, "requires both --frequency and --at"},
		{"target without apply-to", []string{"--target", testID}, "--target requires --apply-to"},
		{"nothing", nil, "nothing to update"},
		{"xlsx-only template without format", []string{"--template", "DEVICE_ONLINE_RATE"}, "pass --format xlsx"},
		{"no from flag", []string{"--from", "2026-10-01"}, "unknown flag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, reqs := recorder(t, map[string]any{})
			_, _, err := run(t, srv.URL, append([]string{"policy", "update", testID}, tt.args...)...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
			}
			if len(*reqs) != 0 {
				t.Errorf("no request expected, got %+v", *reqs)
			}
		})
	}
}

func TestPolicyTableRows(t *testing.T) {
	in := `{"result":[
		{"_id":"1","schedule":{"hourOfDay":9,"minuteOfHour":5},"templates":["A","B"]},
		{"_id":"2","schedule":{"dayOfWeek":1,"hourOfDay":23,"minuteOfHour":0}},
		{"_id":"3","schedule":{"dayOfMonth":15,"hourOfDay":0,"minuteOfHour":30}}]}`
	out, err := policyTableRows([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatal(err)
	}
	want := []struct{ freq, at string }{{"daily", "09:05"}, {"weekly (MONDAY)", "23:00"}, {"monthly (day 15)", "00:30"}}
	for i, w := range want {
		if rows[i]["frequency"] != w.freq || rows[i]["at"] != w.at {
			t.Errorf("row %d = %v, want %+v", i, rows[i], w)
		}
		if _, ok := rows[i]["schedule"]; ok {
			t.Errorf("row %d still has schedule", i)
		}
	}
	if rows[0]["templateCount"] != float64(2) {
		t.Errorf("templates = %v, want 2", rows[0]["templates"])
	}
}

func TestListAliases(t *testing.T) {
	for _, args := range [][]string{{"ls"}, {"policy", "ls"}} {
		srv, reqs := recorder(t, map[string]any{"result": []any{}, "total": 0})
		if _, _, err := run(t, srv.URL, append(args, "--name", "wk", "-o", "table")...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if len(*reqs) != 1 || !strings.Contains((*reqs)[0].query, "name=wk") ||
			!strings.Contains((*reqs)[0].query, "applyTo") {
			t.Errorf("%v: requests = %+v", args, *reqs)
		}
	}
}

func TestDownload(t *testing.T) {
	file := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("presigned download must not carry the API token")
		}
		w.Write([]byte("PDFDATA"))
	}))
	defer file.Close()

	srv, _ := recorder(t, map[string]any{"result": map[string]any{"downloadUrl": file.URL + "/r.pdf"}})

	stdout, _, err := run(t, srv.URL, "download", testID)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stdout) != file.URL+"/r.pdf" {
		t.Errorf("stdout = %q, want the URL", stdout)
	}

	dest := filepath.Join(t.TempDir(), "r.pdf")
	if _, _, err := run(t, srv.URL, "download", testID, "--output-file", dest); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(dest); string(data) != "PDFDATA" {
		t.Errorf("file = %q", data)
	}
}

func TestDownload_NoURL(t *testing.T) {
	srv, _ := recorder(t, map[string]any{"result": map[string]any{}})
	if _, _, err := run(t, srv.URL, "download", testID); err == nil || !strings.Contains(err.Error(), "not be completed") {
		t.Fatalf("err = %v", err)
	}
}

func TestTemplates(t *testing.T) {
	stdout, _, err := run(t, "http://unused.invalid", "policy", "templates", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"TOTAL_DEVICES", "CLIENT_DATA_USAGE_DETAILS"} {
		if !strings.Contains(stdout, name) {
			t.Errorf("output missing %s", name)
		}
	}
}
