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

func newDeviceListRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdDevice(f))
	return root
}

func TestDeviceList_DefaultQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	root.SetArgs([]string{"device", "list", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	// Default: page=0, limit=20, no sort, no fields (json output), no expand
	if !strings.Contains(gotQuery, "page=0") {
		t.Errorf("query %q missing page=0", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=20") {
		t.Errorf("query %q missing limit=20", gotQuery)
	}
	if strings.Contains(gotQuery, "sort=") {
		t.Errorf("query %q should not contain sort when not specified", gotQuery)
	}
	if strings.Contains(gotQuery, "fields=") {
		t.Errorf("query %q should not contain fields for json output", gotQuery)
	}
	if strings.Contains(gotQuery, "expand=") {
		t.Errorf("query %q should not contain expand when not specified", gotQuery)
	}
}

func TestDeviceList_TableDefaultFields(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	// No -o flag → defaults to table → default fields should be applied
	root.SetArgs([]string{"device", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	// Should have default fields for table output
	if !strings.Contains(gotQuery, "fields=") {
		t.Errorf("query %q should contain fields for table output", gotQuery)
	}
	// Check some expected default fields
	for _, field := range []string{"_id", "name", "serialNumber", "online"} {
		if !strings.Contains(gotQuery, field) {
			t.Errorf("query %q missing default field %q", gotQuery, field)
		}
	}
}

func TestDeviceList_WithPagination(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	root.SetArgs([]string{"device", "list", "--page", "3", "--limit", "50", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	// page 3 → API page 2
	if !strings.Contains(gotQuery, "page=2") {
		t.Errorf("query %q missing page=2 (CLI page 3 → API page 2)", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=50") {
		t.Errorf("query %q missing limit=50", gotQuery)
	}
}

func TestDeviceList_WithSort(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	root.SetArgs([]string{"device", "list", "--sort", "name,asc", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	if !strings.Contains(gotQuery, "sort=name") {
		t.Errorf("query %q missing sort=name,asc", gotQuery)
	}
}

func TestDeviceList_WithExpand(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	root.SetArgs([]string{"device", "list", "--expand", "org,firmwareUpgradeStatus", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	if !strings.Contains(gotQuery, "expand=org") {
		t.Errorf("query %q missing expand with org", gotQuery)
	}
}

func TestDeviceList_UserFieldsOverrideDefaults(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	// table output + explicit fields → user fields should override defaults
	root.SetArgs([]string{"device", "list", "-f", "name,online"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	if !strings.Contains(gotQuery, "fields=name") {
		t.Errorf("query %q should contain user-specified fields", gotQuery)
	}
	// Should NOT contain defaultListFields like serialNumber
	if strings.Contains(gotQuery, "serialNumber") {
		t.Errorf("query %q should not contain default fields when user specifies fields", gotQuery)
	}
}

func TestDeviceList_WithFilters(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{},
			"total":  0,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newDeviceListRoot(f)
	root.SetArgs([]string{"device", "list", "--online", "true", "-q", "router", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("device list: %v", err)
	}

	if !strings.Contains(gotQuery, "online=true") {
		t.Errorf("query %q missing online=true", gotQuery)
	}
	if !strings.Contains(gotQuery, "q=router") {
		t.Errorf("query %q missing q=router", gotQuery)
	}
}
