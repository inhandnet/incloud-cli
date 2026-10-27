package device

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newAPRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdAP(f))
	return root
}

func runAPList(t *testing.T, resp string, args ...string) (string, string, error) {
	t.Helper()
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte(resp))
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAPRoot(f)
	root.SetArgs(append([]string{"ap", "list", "507f1f77bcf86cd799439011"}, args...))
	err := root.Execute()
	return gotPath, f.IO.Out.(*bytes.Buffer).String(), err
}

const apListResp = `{"result":{"timestamp":1,"apList":[
 {"mac":"AA:00:00:00:00:09","status":"offline"},
 {"mac":"AA:00:00:00:00:02","level":1,"upperMac":"AA:00:00:00:00:01","staList":[{"mac":"s1"},{"mac":"s2"}]},
 {"mac":"AA:00:00:00:00:01","level":0,"staList":[]}]}}`

func TestAPList_TableSortedByLevelWithStaCount(t *testing.T) {
	path, out, err := runAPList(t, apListResp, "-o", "table")
	if err != nil {
		t.Fatalf("ap list: %v", err)
	}
	if path != "/api/v1/network/devices/507f1f77bcf86cd799439011/ap-list" {
		t.Errorf("path = %q", path)
	}
	if strings.Index(out, "AA:00:00:00:00:01") > strings.Index(out, "AA:00:00:00:00:02") {
		t.Errorf("AC (level 0) should come first:\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "sta") || strings.Contains(out, "s1") {
		t.Errorf("table should show staCount, not staList:\n%s", out)
	}
}

func TestAPList_JSONKeepsStaList(t *testing.T) {
	_, out, err := runAPList(t, apListResp, "-o", "json")
	if err != nil {
		t.Fatalf("ap list: %v", err)
	}
	if !strings.Contains(out, "staList") {
		t.Errorf("json output should keep staList:\n%s", out)
	}
}

func TestAPList_Snapshot(t *testing.T) {
	path, _, err := runAPList(t, apListResp, "--snapshot", "-o", "json")
	if err != nil {
		t.Fatalf("ap list --snapshot: %v", err)
	}
	if !strings.HasSuffix(path, "/ap-list/snapshot") {
		t.Errorf("path = %q, want snapshot endpoint", path)
	}
}

func TestAPList_BusinessErrors(t *testing.T) {
	cases := map[string]struct {
		resp string
		args []string
		want string
	}{
		"unlicensed":     {`{"error":"unlicensed_device","message":"device has no license"}`, nil, "not licensed"},
		"offline":        {`{"error":"device offline","message":"x"}`, nil, "--snapshot"},
		"other":          {`{"error":"request failed","message":"boom"}`, nil, "boom"},
		"empty snapshot": {`{"result":null}`, []string{"--snapshot"}, "no networking snapshot"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, out, err := runAPList(t, c.resp, append(c.args, "-o", "table")...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want containing %q", err, c.want)
			}
			if out != "" {
				t.Errorf("no table should be printed on error, got:\n%s", out)
			}
		})
	}
}
