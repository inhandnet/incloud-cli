package user

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

func newUserRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdUser(f))
	return root
}

func newUserServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]any{"result": []any{}, "total": 0})
	}))
	t.Cleanup(srv.Close)
	return srv, &gotQuery
}

func TestUserList_TableDefaultFields(t *testing.T) {
	srv, gotQuery := newUserServer(t)
	f := newTestFactory(t, srv.URL)
	root := newUserRoot(f)
	root.SetArgs([]string{"user", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("user list: %v", err)
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
