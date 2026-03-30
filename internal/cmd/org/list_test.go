package org

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

func newTestFactory(t *testing.T, host string) *factory.Factory {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := &config.Config{
		CurrentContext: "test",
		Contexts: map[string]*config.Context{
			"test": {Host: host, Token: "test-token"},
		},
	}
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}
	return &factory.Factory{
		IO: &iostreams.IOStreams{
			In:     strings.NewReader(""),
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
		ConfigPath: cfgPath,
	}
}

func newOrgRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdOrg(f))
	return root
}

func newOrgServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"result": []any{}, "total": 0})
	}))
	t.Cleanup(srv.Close)
	return srv, &gotQuery
}

func TestOrgList_DefaultQuery(t *testing.T) {
	srv, gotQuery := newOrgServer(t)
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "list", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org list: %v", err)
	}

	if !strings.Contains(*gotQuery, "page=0") {
		t.Errorf("query %q missing page=0", *gotQuery)
	}
	if !strings.Contains(*gotQuery, "limit=20") {
		t.Errorf("query %q missing limit=20", *gotQuery)
	}
	if strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should not contain fields for json output", *gotQuery)
	}
}

func TestOrgList_TableDefaultFields(t *testing.T) {
	srv, gotQuery := newOrgServer(t)
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org list: %v", err)
	}

	if !strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should contain fields for table output", *gotQuery)
	}
	for _, field := range defaultListFields {
		if !strings.Contains(*gotQuery, field) {
			t.Errorf("query %q missing default field %q", *gotQuery, field)
		}
	}
}

func TestOrgList_ExpandFlag(t *testing.T) {
	srv, gotQuery := newOrgServer(t)
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "list", "--expand", "parent", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org list --expand: %v", err)
	}

	if !strings.Contains(*gotQuery, "expand=parent") {
		t.Errorf("query %q missing expand=parent", *gotQuery)
	}
}

func TestOrgList_SortFlag(t *testing.T) {
	srv, gotQuery := newOrgServer(t)
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "list", "--sort", "name,asc", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org list --sort: %v", err)
	}

	if !strings.Contains(*gotQuery, "sort=name") {
		t.Errorf("query %q missing sort=name", *gotQuery)
	}
}

func TestOrgList_Pagination(t *testing.T) {
	srv, gotQuery := newOrgServer(t)
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "list", "--page", "3", "--limit", "50", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org list --page --limit: %v", err)
	}

	if !strings.Contains(*gotQuery, "page=2") {
		t.Errorf("query %q: page 3 should map to page=2", *gotQuery)
	}
	if !strings.Contains(*gotQuery, "limit=50") {
		t.Errorf("query %q missing limit=50", *gotQuery)
	}
}
