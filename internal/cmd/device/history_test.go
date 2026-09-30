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

	"github.com/inhandnet/incloud-cli/internal/factory"
)

const sampleRecordID = "e95bba0b-1f2e-4d3c-9b8a-7f6e5d4c3b2a"

var sampleRecord = map[string]interface{}{
	"record_id":  sampleRecordID,
	"device_id":  "507f1f77bcf86cd799439011",
	"symptom":    "连续 4 天约 14:45 出现 CONNECTION_LOST",
	"root_cause": "L2TP 隧道到 1.1.1.1:1701 持续超时",
	"fix":        nil,
	"status":     "pending",
	"created_at": "2026-09-28T15:11:00Z",
}

func newHistoryRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdHistory(f))
	return root
}

func historyServer(t *testing.T, records []interface{}, gotQuery *url.Values) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != diagnosisRecordsPath {
			t.Errorf("path = %q, want %q", r.URL.Path, diagnosisRecordsPath)
		}
		*gotQuery = r.URL.Query()
		json.NewEncoder(w).Encode(map[string]interface{}{"records": records, "count": len(records)})
	}))
}

func TestHistoryList_DefaultsToOpen(t *testing.T) {
	var q url.Values
	srv := historyServer(t, []interface{}{sampleRecord}, &q)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "list", "507f1f77bcf86cd799439011", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("history list: %v", err)
	}

	if q.Get("device_id") != "507f1f77bcf86cd799439011" || q.Get("status") != "open" {
		t.Errorf("query = %v, want device_id and status=open", q)
	}
	if q.Get("page") != "0" || q.Get("limit") != "20" {
		t.Errorf("query = %v, want page=0 limit=20", q)
	}
	var got []map[string]interface{}
	if err := json.Unmarshal(f.IO.Out.(*bytes.Buffer).Bytes(), &got); err != nil {
		t.Fatalf("output is not a JSON array: %v", err)
	}
	if len(got) != 1 || got[0]["record_id"] != sampleRecordID {
		t.Errorf("output = %v, want the sample record", got)
	}
}

func TestHistoryList_StatusAll(t *testing.T) {
	var q url.Values
	srv := historyServer(t, []interface{}{}, &q)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "ls", "507f1f77bcf86cd799439011", "--status", "all", "--page", "2"})
	if err := root.Execute(); err != nil {
		t.Fatalf("history ls: %v", err)
	}
	if q.Get("status") != "all" || q.Get("page") != "1" {
		t.Errorf("query = %v, want status=all page=1", q)
	}
}

func TestHistoryList_RejectsUnknownStatus(t *testing.T) {
	f, _ := newTestFactory(t, "http://127.0.0.1:1")
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "list", "507f1f77bcf86cd799439011", "--status", "closed"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "--status") {
		t.Errorf("err = %v, want invalid --status", err)
	}
}

func TestHistoryGet_ShortIDs(t *testing.T) {
	var q url.Values
	srv := historyServer(t, []interface{}{sampleRecord}, &q)
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "get", "#E95BBA0B", "68343905", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("history get: %v", err)
	}

	if got := q["record_ids"]; len(got) != 2 || got[0] != "e95bba0b" || got[1] != "68343905" {
		t.Errorf("record_ids = %v, want normalized short ids", got)
	}
	if !strings.Contains(errBuf.String(), "68343905") {
		t.Errorf("stderr = %q, want the unmatched id reported", errBuf.String())
	}
}

func TestHistoryGet_DropsUnrequestedRecords(t *testing.T) {
	other := map[string]interface{}{"record_id": "218deb1b-5a90-44bc-afee-12281d29428e", "status": "pending"}
	var q url.Values
	srv := historyServer(t, []interface{}{other, sampleRecord}, &q)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "get", "e95bba0b", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("history get: %v", err)
	}
	var got []map[string]interface{}
	if err := json.Unmarshal(f.IO.Out.(*bytes.Buffer).Bytes(), &got); err != nil {
		t.Fatalf("output is not a JSON array: %v", err)
	}
	if len(got) != 1 || got[0]["record_id"] != sampleRecordID {
		t.Errorf("output = %v, want only the requested record", got)
	}
}

func TestHistoryGet_NotFound(t *testing.T) {
	other := map[string]interface{}{"record_id": "218deb1b-5a90-44bc-afee-12281d29428e"}
	var q url.Values
	srv := historyServer(t, []interface{}{other}, &q)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "get", "e95bba0b"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "no diagnosis record") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestHistoryGet_RejectsInvalidIDs(t *testing.T) {
	for _, id := range []string{"e95bba0", "rec-1", "e95bba0b.*", sampleRecordID + "0"} {
		f, _ := newTestFactory(t, "http://127.0.0.1:1")
		root := newHistoryRoot(f)
		root.SetArgs([]string{"history", "get", id})
		if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "invalid record id") {
			t.Errorf("id %q: err = %v, want invalid record id", id, err)
		}
	}
}

func TestHistoryTable_ShortensIDs(t *testing.T) {
	var q url.Values
	srv := historyServer(t, []interface{}{sampleRecord}, &q)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newHistoryRoot(f)
	root.SetArgs([]string{"history", "get", "e95bba0b", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("history get: %v", err)
	}
	out := f.IO.Out.(*bytes.Buffer).String()
	if !strings.Contains(out, "e95bba0b") || strings.Contains(out, sampleRecordID) {
		t.Errorf("table = %q, want the short id only", out)
	}
	if strings.Contains(out, "<nil>") {
		t.Errorf("table = %q, want null fields rendered empty", out)
	}
}
