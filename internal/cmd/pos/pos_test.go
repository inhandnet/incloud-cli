package pos

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
			"test": {Host: host, Token: "test-token"},
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

// newPosRoot builds a root command with the global -o flag and the pos group attached.
func newPosRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdPos(f))
	return root
}

func TestPosSubcommands(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	cmd := NewCmdPos(f)

	want := map[string]bool{
		"clients": false, "forwarded": false, "device-hits": false,
		"marked-clients": false, "vendor-hits": false, "vendor-summary": false,
		"client-types": false, "rules": false,
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

func TestPosClients_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []any{}, "total": 0, "page": 0, "limit": 20,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "clients", "--level", "priority", "--device", "dev1", "--expand", "device,org", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos clients: %v", err)
	}
	if gotPath != "/api/v1/pos-ready/clients" {
		t.Errorf("path = %q", gotPath)
	}
	for _, want := range []string{"level=priority", "deviceId=dev1", "expand=device", "page=0"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected %q in query, got: %s", want, gotQuery)
		}
	}
}

func TestPosRulesList_Alias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}, "total": 0, "page": 0, "limit": 20})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "ls", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos rules ls alias: %v", err)
	}
}

func TestPosForwarded_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}, "total": 0, "page": 0, "limit": 20})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "forwarded", "--active-within", "7d", "--vendor", "Verifone", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos forwarded: %v", err)
	}
	if gotPath != "/api/v1/pos-ready/forwarded" {
		t.Errorf("path = %q", gotPath)
	}
	for _, want := range []string{"activeWithin=7d", "vendor=Verifone"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected %q in query, got: %s", want, gotQuery)
		}
	}
}

func TestPosDeviceHits_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"vendors": []any{}}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "device-hits", "dev1", "--group-by", "client", "--active-within", "6h", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos device-hits: %v", err)
	}
	if gotPath != "/api/v1/network/devices/dev1/pos-hits" {
		t.Errorf("path = %q", gotPath)
	}
	for _, want := range []string{"groupBy=client", "activeWithin=6h"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("expected %q in query, got: %s", want, gotQuery)
		}
	}
}

func TestPosMarkedClients_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "marked-clients", "dev1", "--level", "bypass", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos marked-clients: %v", err)
	}
	if gotPath != "/api/v1/network/devices/dev1/marked-clients" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "level=bypass") {
		t.Errorf("expected level=bypass, got: %s", gotQuery)
	}
}

func TestPosVendorHits_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"series": []any{}}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "vendor-hits", "dev1", "c1",
		"--after", "2026-03-17T00:00:00Z", "--before", "2026-03-18T00:00:00Z", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos vendor-hits: %v", err)
	}
	if gotPath != "/api/v1/network/devices/dev1/clients/c1/pos-vendor-hits" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "after=") || !strings.Contains(gotQuery, "before=") {
		t.Errorf("expected after/before, got: %s", gotQuery)
	}
}

func TestPosVendorHits_RequiresAfterBefore(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "vendor-hits", "dev1", "c1"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing --after/--before")
	}
}

func TestPosClientTypes_Basic(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"clientTypes": []any{}}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "client-types", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos client-types: %v", err)
	}
	if gotPath != "/api/v1/client-identification/client-types" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPosRulesGet_Basic(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"rules": []any{}}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "get", "dev1", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos rules get: %v", err)
	}
	if gotPath != "/api/v1/network/devices/dev1/pos/custom-rules" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestPosRulesList_Basic(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"result": []any{}, "total": 0, "page": 0, "limit": 20})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "list", "--device", "dev1", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos rules list: %v", err)
	}
	if gotPath != "/api/v1/network/pos/custom-rules" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(gotQuery, "deviceId=dev1") {
		t.Errorf("expected deviceId=dev1, got: %s", gotQuery)
	}
}

func TestPosRulesSet_FromArrayFile(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{"rules": []any{}}, "pushOutcome": "pushed",
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.json")
	rulesJSON := `[{"type":"add","clientType":"POS_TERMINAL","vendor":"Verifone","protocol":"tcp","address":"1.2.3.4","port":"443"}]`
	if err := os.WriteFile(rulesPath, []byte(rulesJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	f, errBuf := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "set", "dev1", "--file", rulesPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos rules set: %v", err)
	}
	if gotPath != "/api/v1/network/devices/dev1/pos/custom-rules" {
		t.Errorf("path = %q", gotPath)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	rules, ok := gotBody["rules"].([]any)
	if !ok || len(rules) != 1 {
		t.Errorf("body.rules = %v, want 1 entry", gotBody["rules"])
	}
	if !strings.Contains(errBuf.String(), "push: pushed") {
		t.Errorf("expected push outcome in stderr, got: %s", errBuf.String())
	}
}

func TestPosRulesSet_FromObjectFile(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{"rules": []any{}}, "pushOutcome": "pushed",
		})
	}))
	defer srv.Close()

	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.json")
	rulesJSON := `{"rules":[{"type":"mask","clientType":"OTHER","vendor":"Generic"}]}`
	if err := os.WriteFile(rulesPath, []byte(rulesJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	f, _ := newTestFactory(t, srv.URL)
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "set", "dev1", "--file", rulesPath})
	if err := root.Execute(); err != nil {
		t.Fatalf("pos rules set object file: %v", err)
	}
	rules, ok := gotBody["rules"].([]any)
	if !ok || len(rules) != 1 {
		t.Errorf("body.rules = %v, want 1 entry", gotBody["rules"])
	}
}

func TestPosRulesSet_RequiresFile(t *testing.T) {
	f, _ := newTestFactory(t, "https://example.com")
	root := newPosRoot(f)
	root.SetArgs([]string{"pos", "rules", "set", "dev1"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing --file")
	}
}
