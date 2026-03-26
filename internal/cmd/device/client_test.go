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

// newClientRoot creates a root command with the global -o/--output persistent
// flag and the client subcommand attached. Tests call root.SetArgs and
// root.Execute so that cobra resolves "client <sub> ..." correctly.
func newClientRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdClient(f))
	return root
}

// sampleClient is a realistic API response for a single client.
var sampleClient = map[string]interface{}{
	"_id":       "69b8c537e7f8d2c1e5fffdbc",
	"hostname":  "fc:5c:ee:8c:90:93",
	"name":      "DESKTOP-H2BU344",
	"mac":       "FC:5C:EE:8C:90:93",
	"deviceId":  "68c7bba25f49e702993564ac",
	"online":    false,
	"type":      "wired",
	"ip":        "192.168.2.101",
	"vlan":      1,
	"createdAt": "2026-03-17T03:06:30Z",
}

func TestClientMarkAsset(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 200})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "mark-asset", "c1", "c2"})
	if err := root.Execute(); err != nil {
		t.Fatalf("client mark-asset: %v", err)
	}

	if gotPath != "/api/v1/network/clients/mark-assets" {
		t.Errorf("path = %q, want /api/v1/network/clients/mark-assets", gotPath)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	ids, ok := gotBody["ids"].([]interface{})
	if !ok || len(ids) != 2 {
		t.Errorf("body.ids = %v, want [c1 c2]", gotBody["ids"])
	}
}

var seriesResponse = map[string]interface{}{
	"result": map[string]interface{}{
		"series": []interface{}{
			map[string]interface{}{
				"tags":    map[string]interface{}{},
				"columns": []string{"time", "value"},
				"values":  []interface{}{[]interface{}{"2026-03-17T00:00:00Z", 42}},
			},
		},
	},
}

func jsonHandler(resp interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// --- list ---

func TestClientList_Basic(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{sampleClient},
			"total":  1, "page": 0, "limit": 20,
		})
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "list", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(gotQuery, "page=0") {
		t.Errorf("expected page=0 in query, got: %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "limit=20") {
		t.Errorf("expected limit=20 in query, got: %s", gotQuery)
	}
}

func TestClientList_Filters(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{}, "total": 0, "page": 0, "limit": 5,
		})
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "list",
		"--type", "wireless",
		"--online", "true",
		"--device", "dev123",
		"--mac", "AA:BB:CC",
		"--ip", "192.168.1.1",
		"--limit", "5",
		"--page", "2",
		"-o", "json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{
		"type=wireless", "online=true", "deviceId=dev123",
		"ip=192.168.1.1", "limit=5", "page=1",
	} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected %q in query, got: %s", want, gotQuery)
		}
	}
}

func TestClientList_Alias(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{
		"result": []interface{}{}, "total": 0, "page": 0, "limit": 20,
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "ls", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ls alias should work: %v", err)
	}
}

// --- get ---

func TestClientGet_Basic(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": sampleClient})
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "get", "client123", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/client123" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

func TestClientGet_RequiresArg(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "get"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing arg")
	}
}

func TestClientGet_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"client not found"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "get", "nonexistent", "-o", "json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

// --- update ---

func TestClientUpdate_Basic(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{"_id": "c1", "name": "NewName"},
		})
	}))
	defer server.Close()

	f, errBuf := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "update", "c1", "--name", "NewName"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", gotMethod)
	}
	if gotPath != "/api/v1/network/clients/c1" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if gotBody["name"] != "NewName" {
		t.Errorf("unexpected body: %v", gotBody)
	}
	if !strings.Contains(errBuf.String(), `Client "NewName" (c1) updated.`) {
		t.Errorf("unexpected stderr: %s", errBuf.String())
	}
}

func TestClientUpdate_RequiresName(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "update", "c1"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --name")
	}
}

func TestClientUpdate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid name"}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "update", "c1", "--name", ""})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 in error, got: %v", err)
	}
}

// --- throughput ---

func TestClientThroughput_RequiresAfterBefore(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")

	// No flags at all
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "throughput", "c1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	if !strings.Contains(err.Error(), "after") {
		t.Errorf("expected 'after' in error, got: %v", err)
	}

	// Only --after
	cmd = newClientRoot(f)
	cmd.SetArgs([]string{"client", "throughput", "c1", "--after", "2026-03-17T00:00:00Z"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --before")
	}
	if !strings.Contains(err.Error(), "before") {
		t.Errorf("expected 'before' in error, got: %v", err)
	}
}

func TestClientThroughput_Basic(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(seriesResponse)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "throughput", "c1",
		"--after", "2026-03-17T00:00:00Z", "--before", "2026-03-18T00:00:00Z", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/throughput" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

// --- rssi ---

func TestClientRSSI_RequiresAfterBefore(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "rssi", "c1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	if !strings.Contains(err.Error(), "after") {
		t.Errorf("expected 'after' in error, got: %v", err)
	}
}

func TestClientRSSI_Basic(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(seriesResponse)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "rssi", "c1",
		"--after", "2026-03-17T00:00:00Z", "--before", "2026-03-18T00:00:00Z", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/rssi" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(gotQuery, "after=") || !strings.Contains(gotQuery, "before=") {
		t.Errorf("expected after/before in query, got: %s", gotQuery)
	}
}

// --- sinr ---

func TestClientSINR_RequiresAfterBefore(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "sinr", "c1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	if !strings.Contains(err.Error(), "after") {
		t.Errorf("expected 'after' in error, got: %v", err)
	}
}

// --- datausage-hourly ---

func TestClientDatausageHourly_Basic(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(seriesResponse)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "datausage-hourly", "c1",
		"--after", "2026-03-17T00:00:00Z", "--before", "2026-03-18T00:00:00Z", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/datausage-hourly" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

// --- datausage-daily ---

func TestClientDatausageDaily_Basic(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(seriesResponse)
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "datausage-daily", "c1", "--month", "2026-03", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/datausage-daily" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(gotQuery, "month=2026-03") {
		t.Errorf("expected month param, got: %s", gotQuery)
	}
}

func TestClientDatausageDaily_NoRequiredFlags(t *testing.T) {
	server := httptest.NewServer(jsonHandler(map[string]interface{}{"result": map[string]interface{}{}}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "datausage-daily", "c1", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("datausage-daily should work without flags: %v", err)
	}
}

// --- online-events ---

func TestClientOnlineEvents_Basic(t *testing.T) {
	var gotPath, gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"eventType": "connect", "ssid": "TestSSID",
					"mode": "wireless", "timestamp": "2026-03-17T03:00:00Z",
				},
			},
			"total": 1, "page": 0, "limit": 10,
		})
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "online-events", "c1",
		"--limit", "10", "--after", "2026-03-17T00:00:00Z", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/online-events-list" {
		t.Errorf("unexpected path: %s", gotPath)
	}
	if !strings.Contains(gotQuery, "limit=10") {
		t.Errorf("expected limit=10, got: %s", gotQuery)
	}
	if !strings.Contains(gotQuery, "after=") {
		t.Errorf("expected after param, got: %s", gotQuery)
	}
}

// --- online-stats ---

func TestClientOnlineStats_Basic(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": map[string]interface{}{
				"onlineTime": 10877, "offlineCount": 3, "onlineRate": 0.126,
			},
		})
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	cmd := newClientRoot(f)
	cmd.SetArgs([]string{"client", "online-stats", "c1",
		"--after", "2026-03-17T00:00:00Z", "--before", "2026-03-18T00:00:00Z", "-o", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/v1/network/clients/c1/online-events-chart/statistics" {
		t.Errorf("unexpected path: %s", gotPath)
	}
}

// --- subcommand structure ---

func TestClientSubcommands(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := NewCmdClient(f)

	want := map[string]bool{
		"list": false, "get": false, "update": false,
		"throughput": false, "rssi": false, "sinr": false,
		"datausage-hourly": false, "datausage-daily": false,
		"online-events": false, "online-stats": false,
		"mark-asset": false, "set-pos-ready": false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := want[sub.Name()]; ok {
			want[sub.Name()] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

// --- set-pos-ready ---

func TestClientSetPosReady_Basic(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 200})
	}))
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready", "dev123", "--mac", "FC:5C:EE:8C:90:93", "--enabled"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set-pos-ready: %v", err)
	}

	if gotPath != "/api/v1/network/devices/dev123/clients/pos-ready" {
		t.Errorf("path = %q, want /api/v1/network/devices/dev123/clients/pos-ready", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["mac"] != "FC:5C:EE:8C:90:93" {
		t.Errorf("body.mac = %v, want FC:5C:EE:8C:90:93", gotBody["mac"])
	}
	if gotBody["enabled"] != true {
		t.Errorf("body.enabled = %v, want true", gotBody["enabled"])
	}
	if !strings.Contains(errBuf.String(), "POS Ready enabled") {
		t.Errorf("unexpected stderr: %s", errBuf.String())
	}
}

func TestClientSetPosReady_Disable(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{"code": 200})
	}))
	defer srv.Close()

	f, errBuf := newTestFactory(t, srv.URL)
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready", "dev123", "--mac", "AA:BB:CC:DD:EE:FF", "--enabled=false"})
	if err := root.Execute(); err != nil {
		t.Fatalf("set-pos-ready disable: %v", err)
	}

	if gotBody["enabled"] != false {
		t.Errorf("body.enabled = %v, want false", gotBody["enabled"])
	}
	if !strings.Contains(errBuf.String(), "POS Ready disabled") {
		t.Errorf("unexpected stderr: %s", errBuf.String())
	}
}

func TestClientSetPosReady_RequiresMac(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready", "dev123", "--enabled"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing --mac")
	}
}

func TestClientSetPosReady_RequiresEnabled(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready", "dev123", "--mac", "AA:BB:CC:DD:EE:FF"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing --enabled")
	}
}

func TestClientSetPosReady_RequiresDeviceID(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing device-id")
	}
}

func TestClientSetPosReady_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"device does not support star_pos_ready"}`))
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newClientRoot(f)
	root.SetArgs([]string{"client", "set-pos-ready", "dev123", "--mac", "AA:BB:CC:DD:EE:FF", "--enabled"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 in error, got: %v", err)
	}
}
