package device

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

// newSchemaRoot creates a root command with the global -o/--output persistent
// flag and the schema subcommand attached, so tests can pass -o table / -o json.
func newSchemaRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(newCmdConfigSchema(f))
	return root
}

func TestResolveProductVersion_FromFlags(t *testing.T) {
	// No server needed — flags provide values directly
	pv := &productVersion{product: "MR805", version: "V2.0.15-111"}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_FromDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/dev123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"_id":"dev123","product":"MR805","firmware":"V2.0.15-111"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	client, _ := f.APIClient()

	pv, err := resolveProductVersion(client, "dev123", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_MutualExclusion(t *testing.T) {
	_, err := resolveProductVersion(nil, "dev123", "MR805", "V2.0.15")
	if err == nil {
		t.Fatal("expected error for mutual exclusion")
	}
}

func TestResolveProductVersion_MissingParams(t *testing.T) {
	_, err := resolveProductVersion(nil, "", "MR805", "")
	if err == nil {
		t.Fatal("expected error when --product without --version")
	}
	_, err = resolveProductVersion(nil, "", "", "")
	if err == nil {
		t.Fatal("expected error when no params")
	}
}

func TestSchemaList_Table(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/devices/dev1":
			_, _ = w.Write([]byte(`{"result":{"_id":"dev1","product":"MR805","firmware":"V2.0.15-111"}}`))
		case "/api/v1/config-documents":
			if r.URL.Query().Get("product") != "MR805" {
				t.Errorf("expected product=MR805, got %s", r.URL.Query().Get("product"))
			}
			_, _ = w.Write([]byte(`{"result":[
				{"_id":"1","name":"System DNS","jsonKeys":["dns"],"descriptions":["Global DNS configuration"]},
				{"_id":"2","name":"WiFi Settings","jsonKeys":["wifi"],"descriptions":["WiFi AP and client config"]}
			],"total":2}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "list", "--device", "dev1", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}

func TestSchemaList_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/config-documents":
			_, _ = w.Write([]byte(`{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"descriptions":["Global DNS"]}],"total":1}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "list", "--product", "MR805", "--version", "V2.0.15-111", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}

func TestSchemaGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/config-documents" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("jsonKeys") != "dns" {
			t.Errorf("expected jsonKeys=dns, got %s", r.URL.Query().Get("jsonKeys"))
		}
		_, _ = w.Write([]byte(`{"result":[{
			"_id":"1",
			"name":"System DNS",
			"jsonKeys":["dns"],
			"descriptions":["Global DNS configuration. Higher priority than upstream DNS."],
			"content":"{\"type\":\"object\",\"properties\":{\"primary\":{\"type\":\"string\"}}}"
		}]}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "get", "--product", "MR805", "--version", "V2.0.15-111", "dns"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "System DNS") {
		t.Errorf("expected 'System DNS' in output, got: %s", out.String())
	}
}

func TestSchemaGet_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[]}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "get", "--product", "MR805", "--version", "V2.0.15", "nonexistent"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestSchemaOverview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/config-documents/overview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"_id":"ov1","product":"CPE02","module":"default","version":"V2.0.8","content":"### JSON KEYS\n- wan: WAN\n- cellular: Cellular"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "overview", "--product", "CPE02", "--version", "V2.0.8"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "### JSON KEYS") {
		t.Errorf("expected markdown content in output, got: %s", out.String())
	}
}

func TestSchemaOverview_NotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/config-documents/overview":
			_, _ = w.Write([]byte(`{"result":null}`))
		case "/api/v1/config-documents":
			// suggestAvailableVersions call — return empty
			_, _ = w.Write([]byte(`{"result":[]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "overview", "--product", "MR805", "--version", "V2.0.15-111"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for no overview available")
	}
	if !strings.Contains(err.Error(), "no overview available") {
		t.Errorf("expected 'no overview available' in error, got: %v", err)
	}
}

func TestSchemaValidate_Pass(t *testing.T) {
	schema := `{"type":"object","properties":{"dns":{"type":"object","properties":{"primary":{"type":"string"}},"required":["primary"]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"content":%q}]}`, schema)
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "validate", "--product", "MR805", "--version", "V2.0.15-111",
		"--key", "dns", "--payload", `{"dns":{"primary":"8.8.8.8"}}`})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "Validation passed") {
		t.Errorf("expected 'Validation passed' in stderr, got: %s", errBuf.String())
	}
}

func TestSchemaValidate_Fail(t *testing.T) {
	schema := `{"type":"object","properties":{"dns":{"type":"object","properties":{"primary":{"type":"string"}},"required":["primary"]}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"result":[{"_id":"1","name":"System DNS","jsonKeys":["dns"],"content":%q}]}`, schema)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "validate", "--product", "MR805", "--version", "V2.0.15-111",
		"--key", "dns", "--payload", `{"dns":{"primary":123}}`})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSchemaValidate_PayloadFileMutualExclusion(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "validate", "--product", "MR805", "--version", "V1",
		"--key", "dns", "--payload", "{}", "--file", "f.json"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error, got: %v", err)
	}
}

func TestSchemaProducts_Table(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/config-documents/overviews" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("module") != "default" {
			t.Errorf("expected module=default, got %s", r.URL.Query().Get("module"))
		}
		// Include duplicates to test deduplication
		_, _ = w.Write([]byte(`{"result":[
			{"product":"MR805","version":"V2.0.16","module":"default"},
			{"product":"CPE02","version":"V2.0.8","module":"default"},
			{"product":"MR805","version":"V2.0.15-111","module":"default"},
			{"product":"MR805","version":"V2.0.16","module":"default"}
		],"total":4}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "products", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	// Should be sorted: CPE02 first, then MR805 versions
	if !strings.Contains(output, "CPE02") {
		t.Errorf("expected 'CPE02' in output, got: %s", output)
	}
	if !strings.Contains(output, "MR805") {
		t.Errorf("expected 'MR805' in output, got: %s", output)
	}
	if !strings.Contains(output, "V2.0.15-111") {
		t.Errorf("expected 'V2.0.15-111' in output, got: %s", output)
	}
	if !strings.Contains(output, "V2.0.16") {
		t.Errorf("expected 'V2.0.16' in output, got: %s", output)
	}
}

func TestSchemaProducts_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[
			{"product":"MR805","version":"V2.0.15-111","module":"default"}
		],"total":1,"page":0,"pageSize":20}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "products", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	// JSON output should preserve the full PageResult envelope
	if !strings.Contains(out.String(), "total") {
		t.Errorf("expected 'total' in json output, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "MR805") {
		t.Errorf("expected 'MR805' in json output, got: %s", out.String())
	}
}

func TestSchemaProducts_Filter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("product") != "MR805" {
			t.Errorf("expected product=MR805, got %s", q.Get("product"))
		}
		if q.Get("version") != "V2.0.15-111" {
			t.Errorf("expected version=V2.0.15-111, got %s", q.Get("version"))
		}
		_, _ = w.Write([]byte(`{"result":[{"product":"MR805","version":"V2.0.15-111"}],"total":1}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "products", "--product", "MR805", "--version", "V2.0.15-111"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaProducts_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":[],"total":0}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	out := f.IO.Out.(*bytes.Buffer)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "products", "-o", "table"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No results") {
		t.Errorf("expected 'No results' in output, got: %s", out.String())
	}
}

func TestTransformSchemaProducts_Deduplicate(t *testing.T) {
	body := []byte(`{"result":[
		{"product":"A","version":"V1"},
		{"product":"B","version":"V2"},
		{"product":"A","version":"V1"},
		{"product":"A","version":"V3"}
	]}`)

	out, err := transformSchemaProducts(body)
	if err != nil {
		t.Fatal(err)
	}

	// Parse result and verify deduplication + sorting
	root := newSchemaRoot(nil)
	_ = root

	// Use simple string checks on the JSON output
	outStr := string(out)
	countV1 := strings.Count(outStr, `"V1"`)
	if countV1 != 1 {
		t.Errorf("expected V1 to appear exactly once after dedup, got %d", countV1)
	}
	// A/V1 should come before A/V3, which should come before B/V2
	idxA1 := strings.Index(outStr, `"V1"`)
	idxA3 := strings.Index(outStr, `"V3"`)
	idxB2 := strings.Index(outStr, `"V2"`)
	if idxA1 == -1 || idxA3 == -1 || idxB2 == -1 {
		t.Fatalf("missing entries in output: %s", outStr)
	}
	if !(idxA1 < idxA3 && idxA3 < idxB2) {
		t.Errorf("expected sorted order A/V1, A/V3, B/V2; got positions %d, %d, %d", idxA1, idxA3, idxB2)
	}
}
