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
    "score": 2.5,
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
	f, _, err := run(t, searchHit, &cap, "search", "IPSec VPN", "--model", "ER805", "--limit", "3", "-o", "table")
	if err != nil {
		t.Fatalf("knowledge search: %v", err)
	}
	if cap.Method != "POST" || cap.Path != "/api/v1/knowledge/search" {
		t.Errorf("got %s %s", cap.Method, cap.Path)
	}
	body := string(cap.Body)
	for _, want := range []string{`"query":"IPSec VPN"`, `"model":"ER805"`, `"limit":3`} {
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

func TestSearch_PathFlagRemoved(t *testing.T) {
	_, _, err := run(t, searchHit, nil, "search", "x", "--path", "device_")
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("want unknown flag error, got %v", err)
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
	f, _, err := run(t, `{"products":[{"product_id":"ER605","display_name":"ER605","kind":"model"}]}`, &cap, "browse", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Path != "/api/v1/knowledge/browse" || string(cap.Body) != "{}" {
		t.Errorf("got %s body %s", cap.Path, cap.Body)
	}
	if !strings.Contains(stdoutOf(f).String(), "ER605") {
		t.Errorf("missing product: %q", stdoutOf(f).String())
	}
}

func TestBrowse_ProductOverview(t *testing.T) {
	var cap captured
	resp := `{"product":"DeviceLive","documents":[{"path":"docs/zh/DeviceLive/Manuals/用户手册/DeviceLive用户手册.md","doc_title":"DeviceLive用户手册","document_type":"manual"}],"sections":[{"chunk_id":"c-9","heading_path":"1. 产品概述"}]}`
	f, _, err := run(t, resp, &cap, "browse", "--product", "DeviceLive", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if string(cap.Body) != `{"product":"DeviceLive"}` {
		t.Errorf("request body %s", cap.Body)
	}
	out := stdoutOf(f).String()
	for _, want := range []string{"DeviceLive用户手册", "[manual]", "1. 产品概述", "[c-9]"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %q", want, out)
		}
	}
}

func TestBrowse_DocumentOutlineWithCursor(t *testing.T) {
	var cap captured
	resp := `{"path":"p.md","doc_title":"ER805用户手册_V1.0","sections":[{"chunk_id":"c-1","heading_path":"1 概述"}],"next_cursor":20}`
	f, errBuf, err := run(t, resp, &cap, "browse", "p.md", "--cursor", "10", "-o", "table")
	if err != nil {
		t.Fatal(err)
	}
	if string(cap.Body) != `{"path":"p.md","cursor":10}` {
		t.Errorf("request body %s", cap.Body)
	}
	if !strings.Contains(stdoutOf(f).String(), "1 概述 [c-1]") {
		t.Errorf("outline missing: %q", stdoutOf(f).String())
	}
	if !strings.Contains(errBuf.String(), "--cursor 20") {
		t.Errorf("missing next-page hint: %q", errBuf.String())
	}
}

func TestBrowse_InvalidCombinationsFailBeforeRequest(t *testing.T) {
	for _, args := range [][]string{
		{"browse", "p.md", "--product", "ER805"},
		{"browse", "--cursor", "10"},
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
	resp := `{"text":"正文","source":{"chunk_id":"c-1","path":"p.md","doc_title":"ER805用户手册_V1.0","heading_path":"5 维护 > 5.1 恢复出厂设置","url":"https://example.com/p.md"},"truncated":true,"next_cursor":12000}`
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
	_, _, err := run(t, `{"text":"","source":{"chunk_id":"c-1","doc_title":"d","heading_path":"h","url":null}}`, &cap, "read", "c-1", "--cursor", "12000", "-o", "json")
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
