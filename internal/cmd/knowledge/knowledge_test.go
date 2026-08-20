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
  "results": [{
    "source": "device_er805_用户手册.md",
    "document_id": "doc-1",
    "section_id": "sec-1",
    "heading_path": "ER805 用户手册 > 4 网络 > 4.2 IPSec VPN",
    "document_type": "device",
    "model": "er805",
    "from_fallback": false,
    "score": 2.5,
    "snippet": "IKE（UDP 500）与 ESP 需要在防火墙中放行。"
  }]
}`

func TestSearch_PostsToAgenticEndpoint(t *testing.T) {
	var cap captured
	srv := captureServer(t, &cap, searchHit)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "search", "IPSec VPN", "--model", "er805", "--path", "device_", "--limit", "5", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge search: %v", err)
	}

	if cap.Method != "POST" || cap.Path != "/api/v1/knowledge/agentic/search" {
		t.Errorf("got %s %s, want POST /api/v1/knowledge/agentic/search", cap.Method, cap.Path)
	}
	if !strings.Contains(string(cap.Body), `"query":"IPSec VPN"`) ||
		!strings.Contains(string(cap.Body), `"model":"er805"`) ||
		!strings.Contains(string(cap.Body), `"path":"device_"`) ||
		!strings.Contains(string(cap.Body), `"limit":5`) {
		t.Errorf("request body %s missing expected fields", cap.Body)
	}
	if strings.Contains(string(cap.Body), "rewrite") {
		t.Errorf("request body %s should not contain rewrite", cap.Body)
	}

	out := stdoutOf(f).String()
	if !strings.Contains(out, "ER805 用户手册 > 4 网络 > 4.2 IPSec VPN") {
		t.Errorf("table output missing heading_path: %q", out)
	}
	if !strings.Contains(out, "[ER805] device_er805_用户手册.md") {
		t.Errorf("table output missing model/source meta: %q", out)
	}
	if !strings.Contains(out, "IKE（UDP 500）") {
		t.Errorf("table output missing snippet: %q", out)
	}
}

func TestSearch_StatusHandling(t *testing.T) {
	tests := []struct {
		name       string
		resp       string
		wantStderr string
		wantStdout string
	}{
		{
			name:       "empty",
			resp:       `{"status": "empty", "results": []}`,
			wantStderr: "No results found.",
		},
		{
			name:       "kb_not_ready",
			resp:       `{"status": "kb_not_ready", "results": []}`,
			wantStderr: "Knowledge base is not ready.",
		},
		{
			name:       "failed",
			resp:       `{"status": "failed", "results": []}`,
			wantStderr: "Search failed on the server.",
		},
		{
			name: "model_not_found fallback",
			resp: `{"status": "model_not_found", "results": [{
				"source": "platform_用户手册.md", "document_id": "d", "section_id": "s",
				"heading_path": "平台 > 告警", "document_type": "platform", "model": "default",
				"from_fallback": true, "score": 1.0, "snippet": "webhook 告警"
			}]}`,
			wantStderr: "fallback results",
			wantStdout: "(fallback)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := captureServer(t, nil, tt.resp)
			defer srv.Close()

			f, errBuf := newTestFactory(t, srv.URL)
			root := newKnowledgeRoot(f)
			root.SetArgs([]string{"knowledge", "search", "q", "-o", "table"})
			if err := root.Execute(); err != nil {
				t.Fatalf("knowledge search: %v", err)
			}
			if tt.wantStderr != "" && !strings.Contains(errBuf.String(), tt.wantStderr) {
				t.Errorf("stderr %q missing %q", errBuf.String(), tt.wantStderr)
			}
			if tt.wantStdout != "" && !strings.Contains(stdoutOf(f).String(), tt.wantStdout) {
				t.Errorf("stdout %q missing %q", stdoutOf(f).String(), tt.wantStdout)
			}
		})
	}
}

func TestSearch_RewriteFlagRemoved(t *testing.T) {
	srv := captureServer(t, nil, searchHit)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "search", "q", "--rewrite"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected unknown-flag error for --rewrite, got %v", err)
	}
}

func TestSearch_JSONPassthrough(t *testing.T) {
	srv := captureServer(t, nil, searchHit)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "search", "IPSec VPN", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge search: %v", err)
	}

	out := stdoutOf(f).String()
	for _, want := range []string{`"status"`, `"document_id":"doc-1"`, `"section_id":"sec-1"`, `"snippet"`} {
		if !strings.Contains(out, want) {
			t.Errorf("json output missing %s: %q", want, out)
		}
	}
}

const browseTree = `{
  "document_id": "doc-1",
  "title": "ER805 用户手册",
  "nodes": [
    {"section_id": "sec-1", "title": "4 网络", "level": 1, "char_count": 800, "child_count": 2},
    {"section_id": "sec-2", "title": "4.2 IPSec VPN", "level": 2, "char_count": 300, "child_count": 0}
  ]
}`

func TestBrowse_OutlineTree(t *testing.T) {
	var cap captured
	srv := captureServer(t, &cap, browseTree)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "browse", "doc-1", "--section", "sec-1", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge browse: %v", err)
	}

	if cap.Path != "/api/v1/knowledge/agentic/browse" {
		t.Errorf("got path %s", cap.Path)
	}
	if !strings.Contains(string(cap.Body), `"document_id":"doc-1"`) ||
		!strings.Contains(string(cap.Body), `"section_id":"sec-1"`) {
		t.Errorf("request body %s missing ids", cap.Body)
	}

	out := stdoutOf(f).String()
	if !strings.Contains(out, "ER805 用户手册") {
		t.Errorf("output missing title: %q", out)
	}
	if !strings.Contains(out, "4 网络") || !strings.Contains(out, "4.2 IPSec VPN") {
		t.Errorf("output missing nodes: %q", out)
	}
	if !strings.Contains(out, "[sec-2]") {
		t.Errorf("output missing section id for follow-up read: %q", out)
	}
	indented := strings.Index(out, "4.2 IPSec VPN") - strings.Index(out, "4 网络")
	if indented <= 0 {
		t.Errorf("child node should be indented after parent: %q", out)
	}
}

func TestBrowse_DocumentNotFound(t *testing.T) {
	srv := captureServer(t, nil, `{"document_id": "nope", "title": "", "nodes": []}`)
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "browse", "nope", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge browse: %v", err)
	}
	if !strings.Contains(errBuf.String(), "Document not found") {
		t.Errorf("stderr %q missing not-found hint", errBuf.String())
	}
}

func TestGrep_MatchesAndRequestShape(t *testing.T) {
	var cap captured
	srv := captureServer(t, &cap, `{
  "pattern": "IKE", "match_count": 2, "truncated": true,
  "matches": [
    {"document_id": "d", "section_id": "s1", "heading_path": "A > B", "line": 42, "text": "IKE UDP 500"},
    {"document_id": "d", "section_id": "s2", "heading_path": "A > C", "line": 7, "text": "ike rekey"}
  ]
}`)
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "grep", "IKE", "--doc", "d", "--path", "device_", "--limit", "2", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge grep: %v", err)
	}

	if cap.Path != "/api/v1/knowledge/agentic/grep" {
		t.Errorf("got path %s", cap.Path)
	}
	// CLI 默认大小写敏感：ignore_case 显式为 false（不依赖服务端默认 true）
	if !strings.Contains(string(cap.Body), `"ignore_case":false`) {
		t.Errorf("request body %s should pin ignore_case=false", cap.Body)
	}
	if !strings.Contains(string(cap.Body), `"document_id":"d"`) {
		t.Errorf("request body %s missing document_id", cap.Body)
	}

	out := stdoutOf(f).String()
	if !strings.Contains(out, "A > B : 42 : IKE UDP 500") {
		t.Errorf("output missing grep-style match line: %q", out)
	}
	if !strings.Contains(errBuf.String(), "truncated") {
		t.Errorf("stderr %q missing truncated hint", errBuf.String())
	}
}

func TestGrep_IgnoreCaseFlagAndNoMatches(t *testing.T) {
	var cap captured
	srv := captureServer(t, &cap, `{"pattern": "x", "match_count": 0, "matches": [], "truncated": false}`)
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "grep", "x", "-i", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge grep: %v", err)
	}

	if !strings.Contains(string(cap.Body), `"ignore_case":true`) {
		t.Errorf("request body %s should carry ignore_case=true with -i", cap.Body)
	}
	if !strings.Contains(errBuf.String(), "No matches found.") {
		t.Errorf("stderr %q missing no-matches hint", errBuf.String())
	}
}

const readBody = `{
  "text": "IKE（UDP 500）与 ESP（协议号 50）需要在防火墙中放行。",
  "source": {
    "document_id": "doc-1", "section_id": "sec-1", "s3_key": "cn/device_er805.md",
    "title": "ER805 用户手册", "heading_path": "ER805 用户手册 > 4 网络 > 4.2 IPSec VPN"
  },
  "truncated": false
}`

func TestRead_DualIDAndPagination(t *testing.T) {
	var cap captured
	srv := captureServer(t, &cap, readBody)
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "read", "sec-1", "--offset", "10", "--limit", "50", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge read: %v", err)
	}

	if cap.Path != "/api/v1/knowledge/agentic/read" {
		t.Errorf("got path %s", cap.Path)
	}
	// 双字段消歧：同一个位置参数同时作为 section_id 与 document_id 发送
	if !strings.Contains(string(cap.Body), `"section_id":"sec-1"`) ||
		!strings.Contains(string(cap.Body), `"document_id":"sec-1"`) {
		t.Errorf("request body %s should carry the arg as both ids", cap.Body)
	}
	if !strings.Contains(string(cap.Body), `"offset":10`) ||
		!strings.Contains(string(cap.Body), `"limit":50`) {
		t.Errorf("request body %s missing pagination", cap.Body)
	}

	out := stdoutOf(f).String()
	if !strings.Contains(out, "IKE（UDP 500）") {
		t.Errorf("output missing text: %q", out)
	}
	if !strings.Contains(out, "[source: ER805 用户手册 > ER805 用户手册 > 4 网络 > 4.2 IPSec VPN]") {
		t.Errorf("output missing source meta: %q", out)
	}
}

func TestRead_NotFoundAndTruncatedHint(t *testing.T) {
	srv := captureServer(t, nil, `{"text": "", "source": null, "truncated": false}`)
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newKnowledgeRoot(f)
	root.SetArgs([]string{"knowledge", "read", "nope", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("knowledge read: %v", err)
	}
	if !strings.Contains(errBuf.String(), "not found") {
		t.Errorf("stderr %q missing not-found hint", errBuf.String())
	}

	srv2 := captureServer(t, nil, `{"text": "…", "source": {"document_id": "d", "section_id": "", "s3_key": "k", "title": "T", "heading_path": "H"}, "truncated": true}`)
	defer srv2.Close()
	f2, errBuf2 := newTestFactory(t, srv2.URL)
	root2 := newKnowledgeRoot(f2)
	root2.SetArgs([]string{"knowledge", "read", "d", "-o", "table"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("knowledge read: %v", err)
	}
	if !strings.Contains(errBuf2.String(), "--offset") {
		t.Errorf("stderr %q missing truncation hint", errBuf2.String())
	}
}
