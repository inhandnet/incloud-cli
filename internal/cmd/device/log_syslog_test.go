package device

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type syslogRun struct {
	stdout string
	stderr string
	query  url.Values
}

func runLogSyslog(t *testing.T, resp string, args ...string) syslogRun {
	t.Helper()

	var query url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	out := &bytes.Buffer{}
	f.IO.Out = out

	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	root.SetArgs(append([]string{
		"device", "log", "syslog", "507f1f77bcf86cd799439011",
		"--after", "2026-11-19T00:00:00Z", "--before", "2026-11-20T00:00:00Z",
	}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("device log syslog %v: %v", args, err)
	}
	return syslogRun{stdout: out.String(), stderr: errBuf.String(), query: query}
}

const syslogTwoLines = `{"total":5,"result":[` +
	`"<30> Nov 19 08:44:22 syswatcher[1]: bootflag is 0xFFFFFFFC\n",` +
	`"<30> Nov 19 10:48:02 conf_proc[1043]: done\n"]}`

func TestLogSyslogTruncatedHintOnStderr(t *testing.T) {
	r := runLogSyslog(t, syslogTwoLines, "--limit", "2")

	if strings.Contains(r.stdout, "Showing") {
		t.Errorf("truncation note leaked into stdout: %q", r.stdout)
	}
	if got := strings.Count(r.stdout, "\n"); got != 2 {
		t.Errorf("stdout should hold the 2 log lines, got %d: %q", got, r.stdout)
	}
	for _, want := range []string{"Showing lines 1-2 of 5", `last line at "Nov 19 10:48:02"`, "--page 2"} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("stderr missing %q: %q", want, r.stderr)
		}
	}
	if r.query.Get("index") != "0" || r.query.Get("limit") != "2" {
		t.Errorf("query index=%q limit=%q, want 0 and 2", r.query.Get("index"), r.query.Get("limit"))
	}
}

func TestLogSyslogNoHintWhenComplete(t *testing.T) {
	resp := `{"total":2,"result":["<30> Nov 19 08:44:22 a\n","<30> Nov 19 08:44:23 b\n"]}`
	r := runLogSyslog(t, resp)
	if r.stderr != "" {
		t.Errorf("expected no stderr output, got %q", r.stderr)
	}
}

func TestLogSyslogNoHintWhenEmpty(t *testing.T) {
	r := runLogSyslog(t, `{"total":0,"result":[]}`)
	if r.stderr != "" || r.stdout != "" {
		t.Errorf("expected no output, got stdout=%q stderr=%q", r.stdout, r.stderr)
	}
}

func TestLogSyslogPage(t *testing.T) {
	r := runLogSyslog(t, syslogTwoLines, "--limit", "2", "--page", "2")

	if r.query.Get("index") != "1" {
		t.Errorf("--page 2 should send index=1, got %q", r.query.Get("index"))
	}
	for _, want := range []string{"Showing lines 3-4 of 5", "--page 3"} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("stderr missing %q: %q", want, r.stderr)
		}
	}
}

func TestLogSyslogLastPageNoHint(t *testing.T) {
	resp := `{"total":5,"result":["<30> Nov 19 11:00:00 last\n"]}`
	r := runLogSyslog(t, resp, "--limit", "2", "--page", "3")
	if r.stderr != "" {
		t.Errorf("last page should not print a hint, got %q", r.stderr)
	}
}

func TestLogSyslogHintWithoutTimestamp(t *testing.T) {
	resp := `{"total":3,"result":["no header here\n"]}`
	r := runLogSyslog(t, resp, "--limit", "1")
	if !strings.Contains(r.stderr, "Showing lines 1-1 of 3.") || strings.Contains(r.stderr, "last line at") {
		t.Errorf("unexpected stderr: %q", r.stderr)
	}
}

func TestLogSyslogJSON(t *testing.T) {
	r := runLogSyslog(t, syslogTwoLines, "--limit", "2", "-o", "json")

	var got syslogResponse
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON (%v): %q", err, r.stdout)
	}
	if got.Total != 5 || len(got.Result) != 2 {
		t.Errorf("got total=%d lines=%d, want 5 and 2", got.Total, len(got.Result))
	}
	if !strings.Contains(r.stderr, "Showing lines 1-2 of 5") {
		t.Errorf("stderr should still carry the note: %q", r.stderr)
	}
}

func TestLogSyslogJSONEmptyResultIsArray(t *testing.T) {
	r := runLogSyslog(t, `{"total":0}`, "-o", "json")
	if !strings.Contains(r.stdout, `"result": []`) {
		t.Errorf("empty result should encode as [], got %q", r.stdout)
	}
}

func TestLogSyslogRejectsBadPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("request should not be sent")
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	root.SetArgs([]string{
		"device", "log", "syslog", "507f1f77bcf86cd799439011",
		"--after", "2026-11-19T00:00:00Z", "--before", "2026-11-20T00:00:00Z", "--page", "0",
	})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "--page") {
		t.Errorf("expected --page validation error, got %v", err)
	}
}
