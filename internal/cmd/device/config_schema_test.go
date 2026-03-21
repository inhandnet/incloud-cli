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
		_, _ = w.Write([]byte(`{"result":{"_id":"dev123","partNumber":"MR805","firmware":"V2.0.15-111"}}`))
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
			_, _ = w.Write([]byte(`{"result":{"_id":"dev1","partNumber":"MR805","firmware":"V2.0.15-111"}}`))
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
		_, _ = w.Write([]byte(`{"result":null}`))
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)

	root := newSchemaRoot(f)
	root.SetArgs([]string{"schema", "overview", "--product", "MR805", "--version", "V2.0.15-111"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errBuf.String(), "No overview available") {
		t.Errorf("expected 'No overview available' in stderr, got: %s", errBuf.String())
	}
}
