package alert

import (
	"bytes"
	"encoding/json"
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

func newAlertRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdAlert(f))
	return root
}

func TestAlertList_TableDefaultFields(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{
			"result": []any{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAlertRoot(f)
	root.SetArgs([]string{"alert", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("alert list: %v", err)
	}

	if !strings.Contains(gotQuery, "fields=") {
		t.Errorf("query %q should contain fields for table output", gotQuery)
	}
	for _, field := range []string{"_id", "type", "priority", "status"} {
		if !strings.Contains(gotQuery, field) {
			t.Errorf("query %q missing default field %q", gotQuery, field)
		}
	}
}

func TestAlertList_WithFilters(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{
			"result": []any{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAlertRoot(f)
	root.SetArgs([]string{"alert", "list", "--status", "ACTIVE", "--ack", "false", "-q", "router", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("alert list: %v", err)
	}

	if !strings.Contains(gotQuery, "status=ACTIVE") {
		t.Errorf("query %q missing status=ACTIVE", gotQuery)
	}
	if !strings.Contains(gotQuery, "ack=false") {
		t.Errorf("query %q missing ack=false", gotQuery)
	}
	if !strings.Contains(gotQuery, "entityName=router") {
		t.Errorf("query %q missing entityName=router", gotQuery)
	}
}
