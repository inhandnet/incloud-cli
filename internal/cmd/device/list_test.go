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

func TestDeviceList_NewFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantKeys []string
	}{
		{
			name:     "org filter",
			args:     []string{"device", "list", "--org", "org123", "-o", "json"},
			wantKeys: []string{"oid=org123"},
		},
		{
			name:     "firmware filter",
			args:     []string{"device", "list", "--firmware", "2.0.0", "-o", "json"},
			wantKeys: []string{"firmware=2.0.0"},
		},
		{
			name:     "name filter",
			args:     []string{"device", "list", "--name", "my-router", "-o", "json"},
			wantKeys: []string{"name=my-router"},
		},
		{
			name:     "serial-number filter",
			args:     []string{"device", "list", "--serial-number", "IR6151234567890", "-o", "json"},
			wantKeys: []string{"serial_number=IR6151234567890"},
		},
		{
			name:     "ip filter",
			args:     []string{"device", "list", "--ip", "192.168.1.1", "-o", "json"},
			wantKeys: []string{"ip=192.168.1.1"},
		},
		{
			name:     "label filter",
			args:     []string{"device", "list", "--label", "env=prod", "--label", "region=us", "-o", "json"},
			wantKeys: []string{"labels=env%3Dprod", "labels=region%3Dus"},
		},
		{
			name:     "iccid filter",
			args:     []string{"device", "list", "--iccid", "89860000000000000000", "-o", "json"},
			wantKeys: []string{"iccid=89860000000000000000"},
		},
		{
			name:     "mac filter",
			args:     []string{"device", "list", "--mac", "00:11:22:33:44:55", "-o", "json"},
			wantKeys: []string{"mac=00%3A11%3A22%3A33%3A44%3A55"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("device list: %v", err)
			}

			for _, key := range tc.wantKeys {
				if !strings.Contains(gotQuery, key) {
					t.Errorf("query %q missing %q", gotQuery, key)
				}
			}
		})
	}
}
