package knowledge

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/config"
	"github.com/inhandnet/incloud-cli/internal/factory"
	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

func newTestFactory(t *testing.T, host string) (*factory.Factory, *bytes.Buffer) {
	t.Helper()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {
				Host:  host,
				Token: "test-token",
			},
		},
	}
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	errBuf := &bytes.Buffer{}
	f := &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: errBuf,
		},
		ConfigPath: cfgPath,
	}
	return f, errBuf
}

func newKnowledgeRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdKnowledge(f))
	return root
}

func stdoutOf(f *factory.Factory) *bytes.Buffer {
	return f.IO.Out.(*bytes.Buffer)
}

// captured records the request the CLI actually sent.
type captured struct {
	Method string
	Path   string
	Body   []byte
}

func captureServer(t *testing.T, cap *captured, resp string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cap != nil {
			cap.Method = r.Method
			cap.Path = r.URL.Path
			cap.Body, _ = io.ReadAll(r.Body)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, resp)
	}))
}

const searchHit = `{
  "status": "success",
  "message": null,
  "results": [{
    "chunk_id": "c-1",
    "path": "docs/zh/ER805/Manuals/用户手册/ER805用户手册_V1.0.md",
    "doc_title": "ER805用户手册_V1.0",
    "heading_path": "ER805用户手册 > 4 网络 > 4.2 IPSec VPN",
    "product_ids": ["ER805"],
    "snippet": "IKE（UDP 500）与 ESP 需要在防火墙中放行。"
  }]
}`

func run(t *testing.T, srvResp string, cap *captured, args ...string) (*factory.Factory, *bytes.Buffer, error) {
	t.Helper()
	srv := captureServer(t, cap, srvResp)
	t.Cleanup(srv.Close)
	f, errBuf := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs(append([]string{"knowledge"}, args...))
	return f, errBuf, root.Execute()
}

func TestKnowledge_HasNoGrep(t *testing.T) {
	f, _ := newTestFactory(t, "http://127.0.0.1")
	for _, c := range NewCmdKnowledge(f).Commands() {
		if c.Name() == "grep" {
			t.Fatal("grep subcommand should be removed")
		}
	}
}

func TestSearch_RequestAndTable(t *testing.T) {
	var cap captured
	f, _, err := run(t, searchHit, &cap, "search", "ER805 IPSec VPN", "--limit", "3", "-o", "table")
	if err != nil {
		t.Fatalf("knowledge search: %v", err)
	}
	if cap.Method != "POST" || cap.Path != "/api/v1/knowledge/search" {
		t.Errorf("got %s %s", cap.Method, cap.Path)
	}
	body := string(cap.Body)
	for _, want := range []string{`"query":"ER805 IPSec VPN"`, `"limit":3`} {
		if !strings.Contains(body, want) {
			t.Errorf("request body %s missing %s", body, want)
		}
	}
	if strings.Contains(body, `"path"`) {
		t.Errorf("request body %s should not carry path", body)
	}
	out := stdoutOf(f).String()
	for _, want := range []string{"ER805用户手册 > 4 网络 > 4.2 IPSec VPN", "[ER805] docs/zh/ER805", "[c-1]", "IKE（UDP 500）"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q: %q", want, out)
		}
	}
}

func TestSearch_PathAndModelFlagsRemoved(t *testing.T) {
	for _, flag := range []string{"--path", "--model"} {
		_, _, err := run(t, searchHit, nil, "search", "x", flag, "y")
		if err == nil || !strings.Contains(err.Error(), "unknown flag") {
			t.Fatalf("%s: want unknown flag error, got %v", flag, err)
		}
	}
}

func TestSearch_FailedAndEmpty(t *testing.T) {
	_, errBuf, err := run(t, `{"status":"failed","message":"documents-mcp 请求过于频繁（429）","results":[]}`, nil, "search", "x", "-o", "table")
	if err != nil || !strings.Contains(errBuf.String(), "Search failed: documents-mcp 请求过于频繁（429）") {
		t.Errorf("failed status: err=%v stderr=%q", err, errBuf.String())
	}
	_, errBuf, err = run(t, `{"status":"empty","results":[]}`, nil, "search", "x", "-o", "table")
	if err != nil || !strings.Contains(errBuf.String(), "No results found.") {
		t.Errorf("empty status: err=%v stderr=%q", err, errBuf.String())
	}
}

func TestSearch_JSONPassthrough(t *testing.T) {
	f, _, err := run(t, searchHit, nil, "search", "x", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdoutOf(f).String(), `"chunk_id":"c-1"`) {
		t.Errorf("json output missing chunk_id: %s", stdoutOf(f).String())
	}
}

func TestBrowse_Products(t *testing.T) {
	var cap captured
	resp := `{"status":"success","products":[{"product_id":"ER605","display_name":"ER605"},{"product_id":"EAGLE-ENERGY-MANAGEMENT","display_name":"白鹰能源管家"}]}`
	f, _, err := run(t, resp, &cap, "browse", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/api/v1/knowledge/browse" || string(cap.Body) != "{}" {
		t.Errorf("got %s body %s", cap.Path, cap.Body)
	}
	out := stdoutOf(f).String()
	if !strings.Contains(out, "ER605\n") || strings.Contains(out, "(ER605)") {
		t.Errorf("display name equal to id should not repeat: %q", out)
	}
	if !strings.Contains(out, "EAGLE-ENERGY-MANAGEMENT (白鹰能源管家)") {
		t.Errorf("missing distinct display name: %q", out)
	}
}

func TestBrowse_ProductOverview(t *testing.T) {
	var cap captured
	resp := `{"status":"success","product":"DeviceLive","sections":[{"chunk_id":"c-9","heading_path":"1. 产品概述","path":"docs/zh/DeviceLive/Manuals/用户手册/DeviceLive用户手册.md"}]}`
	f, _, err := run(t, resp, &cap, "browse", "--product", "DeviceLive", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if string(cap.Body) != `{"product":"DeviceLive"}` {
		t.Errorf("request body %s", cap.Body)
	}
	out := stdoutOf(f).String()
	for _, want := range []string{"1. 产品概述", "docs/zh/DeviceLive/Manuals/用户手册/DeviceLive用户手册.md", "[c-9]"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %q", want, out)
		}
	}
}

func TestBrowse_DocumentOutlineIsFullAndIndented(t *testing.T) {
	var cap captured
	resp := `{"status":"success","sections":[{"chunk_id":"c-1","title":"ER805用户手册","level":1},{"chunk_id":"c-2","title":"5 维护","level":2},{"chunk_id":"c-3","title":"5.1 恢复出厂设置","level":3}]}`
	f, _, err := run(t, resp, &cap, "browse", "p.md", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if string(cap.Body) != `{"path":"p.md"}` {
		t.Errorf("request body %s", cap.Body)
	}
	want := "ER805用户手册 [c-1]\n  5 维护 [c-2]\n    5.1 恢复出厂设置 [c-3]\n"
	if stdoutOf(f).String() != want {
		t.Errorf("outline = %q, want %q", stdoutOf(f).String(), want)
	}
}

func TestBrowse_InvalidUsageFailsBeforeRequest(t *testing.T) {
	for _, args := range [][]string{
		{"browse", "p.md", "--product", "ER805"},
		{"browse", "p.md", "--cursor", "10"},
	} {
		var cap captured
		_, _, err := run(t, "{}", &cap, args...)
		if err == nil {
			t.Errorf("%v: want error", args)
		}
		if cap.Method != "" {
			t.Errorf("%v: request should not be sent", args)
		}
	}
}

func TestRead_RequestTableAndCursorHint(t *testing.T) {
	var cap captured
	resp := `{"text":"正文","source":{"path":"p.md","doc_title":"ER805用户手册_V1.0","heading_path":"5 维护 > 5.1 恢复出厂设置","url":"https://example.com/p.md"},"truncated":true,"next_cursor":12000}`
	f, errBuf, err := run(t, resp, &cap, "read", "c-1", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/api/v1/knowledge/read" || string(cap.Body) != `{"chunk_id":"c-1"}` {
		t.Errorf("got %s body %s", cap.Path, cap.Body)
	}
	out := stdoutOf(f).String()
	if !strings.Contains(out, "正文") || !strings.Contains(out, "[source: ER805用户手册_V1.0 > 5 维护 > 5.1 恢复出厂设置] https://example.com/p.md") {
		t.Errorf("table output: %q", out)
	}
	if !strings.Contains(errBuf.String(), "--cursor 12000") {
		t.Errorf("missing cursor hint: %q", errBuf.String())
	}
}

func TestRead_CursorSentAndLineFlagsGone(t *testing.T) {
	var cap captured
	_, _, err := run(t, `{"text":"","source":{"doc_title":"d","heading_path":"h","url":null}}`, &cap, "read", "c-1", "--cursor", "12000", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	if string(cap.Body) != `{"chunk_id":"c-1","cursor":12000}` {
		t.Errorf("request body %s", cap.Body)
	}
	_, _, err = run(t, "{}", nil, "read", "c-1", "--mode", "range")
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("want unknown flag for --mode, got %v", err)
	}
}

func TestKnowledge_OldServerGivesUpgradeHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"detail":"Not Found"}`)
	}))
	defer srv.Close()
	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "search", "vpn"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "older than this CLI") {
		t.Fatalf("want upgrade hint, got %v", err)
	}
}

func TestBrowseAndRead_EmptyResultIsNotAnError(t *testing.T) {
	cases := []struct {
		resp string
		args []string
		hint string
	}{
		{`{"status":"empty","product":"NOPE","sections":[]}`, []string{"browse", "--product", "NOPE", "-o", "table"}, "Nothing found."},
		{`{"status":"empty","path":"nope.md","sections":[]}`, []string{"browse", "nope.md", "-o", "table"}, "Nothing found."},
		{`{"status":"empty","text":"","truncated":false}`, []string{"read", "stale", "-o", "table"}, "search again"},
	}
	for _, tc := range cases {
		f, errBuf, err := run(t, tc.resp, nil, tc.args...)
		if err != nil {
			t.Fatalf("%v: want no error, got %v", tc.args, err)
		}
		if !strings.Contains(errBuf.String(), tc.hint) || stdoutOf(f).Len() != 0 {
			t.Errorf("%v: stderr=%q stdout=%q", tc.args, errBuf.String(), stdoutOf(f).String())
		}
	}
	f, _, err := run(t, `{"status":"empty","text":"","truncated":false}`, nil, "read", "stale", "-o", "json")
	if err != nil || !strings.Contains(stdoutOf(f).String(), `"status":"empty"`) {
		t.Errorf("json passthrough: err=%v out=%q", err, stdoutOf(f).String())
	}
}

func TestBrowseAndRead_FailedStatusIsNotAnError(t *testing.T) {
	failed := `{"status":"failed","message":"documents-mcp 请求过于频繁（429）"}`
	for _, args := range [][]string{{"browse", "-o", "table"}, {"read", "c-1", "-o", "table"}} {
		f, errBuf, err := run(t, failed, nil, args...)
		if err != nil {
			t.Fatalf("%v: want no error, got %v", args, err)
		}
		if !strings.Contains(errBuf.String(), "failed: documents-mcp 请求过于频繁（429）") || stdoutOf(f).Len() != 0 {
			t.Errorf("%v: stderr=%q stdout=%q", args, errBuf.String(), stdoutOf(f).String())
		}
	}
}
