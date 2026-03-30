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
