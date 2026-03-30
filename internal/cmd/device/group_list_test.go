package device

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/inhandnet/incloud-cli/internal/factory"
)

func newGroupListRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	return root
}

func newGroupServer(t *testing.T) (*httptest.Server, *string) {
	t.Helper()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &gotQuery
}

// TestGroupList_TableDefaultFields verifies that table output injects defaultGroupListFields.
func TestGroupList_TableDefaultFields(t *testing.T) {
	srv, gotQuery := newGroupServer(t)
	f, _ := newTestFactory(t, srv.URL)
	root := newGroupListRoot(f)
	root.SetArgs([]string{"device", "group", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("group list: %v", err)
	}

	if !strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should contain fields for table output", *gotQuery)
	}
	for _, field := range defaultGroupListFields {
		if !strings.Contains(*gotQuery, field) {
			t.Errorf("query %q missing default field %q", *gotQuery, field)
		}
	}
	// Summary fields (online/offline/total) should NOT be present
	for _, f := range []string{"online", "offline", "total"} {
		if strings.Contains(*gotQuery, f) {
			t.Errorf("query %q should not contain summary field %q without --summary", *gotQuery, f)
		}
	}
}

// TestGroupList_SummaryDefaultFields verifies that --summary switches to defaultGroupListSummaryFields.
func TestGroupList_SummaryDefaultFields(t *testing.T) {
	srv, gotQuery := newGroupServer(t)
	f, _ := newTestFactory(t, srv.URL)
	root := newGroupListRoot(f)
	root.SetArgs([]string{"device", "group", "list", "--summary"})
	if err := root.Execute(); err != nil {
		t.Fatalf("group list --summary: %v", err)
	}

	if !strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should contain fields for table+summary output", *gotQuery)
	}
	for _, field := range defaultGroupListSummaryFields {
		if !strings.Contains(*gotQuery, field) {
			t.Errorf("query %q missing summary field %q", *gotQuery, field)
		}
	}
}

// TestGroupList_JSONNoDefaultFields verifies that JSON output skips defaultFields injection.
func TestGroupList_JSONNoDefaultFields(t *testing.T) {
	srv, gotQuery := newGroupServer(t)
	f, _ := newTestFactory(t, srv.URL)
	root := newGroupListRoot(f)
	root.SetArgs([]string{"device", "group", "list", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("group list -o json: %v", err)
	}

	if strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should not contain fields for json output", *gotQuery)
	}
}

// TestGroupList_SummaryJSONNoDefaultFields verifies --summary with JSON also skips fields.
func TestGroupList_SummaryJSONNoDefaultFields(t *testing.T) {
	srv, gotQuery := newGroupServer(t)
	f, _ := newTestFactory(t, srv.URL)
	root := newGroupListRoot(f)
	root.SetArgs([]string{"device", "group", "list", "--summary", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("group list --summary -o json: %v", err)
	}

	if strings.Contains(*gotQuery, "fields=") {
		t.Errorf("query %q should not contain fields for json output even with --summary", *gotQuery)
	}
}

// TestGroupList_UserFieldsOverrideDefaults verifies explicit --fields overrides both default sets.
func TestGroupList_UserFieldsOverrideDefaults(t *testing.T) {
	srv, gotQuery := newGroupServer(t)
	f, _ := newTestFactory(t, srv.URL)
	root := newGroupListRoot(f)
	root.SetArgs([]string{"device", "group", "list", "-f", "name,product"})
	if err := root.Execute(); err != nil {
		t.Fatalf("group list -f: %v", err)
	}

	if !strings.Contains(*gotQuery, "fields=name") {
		t.Errorf("query %q should contain user-specified fields", *gotQuery)
	}
	// Should not contain fields only in the default sets (e.g. createdAt)
	if strings.Contains(*gotQuery, "createdAt") {
		t.Errorf("query %q should not contain default-only field createdAt when user specifies fields", *gotQuery)
	}
}
