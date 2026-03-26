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

func newAssetRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "root", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringP("output", "o", "", "Output format")
	root.AddCommand(NewCmdAsset(f))
	return root
}

var sampleAsset = map[string]interface{}{
	"_id":                "67c8d537e7f8d2c1e5fffdaa",
	"name":              "Office Router",
	"mac":               "00:18:05:AB:CD:EF",
	"number":            "AST-001",
	"category":          "router",
	"status":            "in_use",
	"warrantyExpiration": "2027-12-31",
	"notes":             "2nd floor",
	"createdAt":         "2026-03-01T00:00:00Z",
	"updatedAt":         "2026-03-15T00:00:00Z",
}

func TestAssetList(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": []interface{}{sampleAsset},
			"total":  1,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "list", "--name", "Office", "--category", "router", "--status", "in_use", "-o", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset list: %v", err)
	}

	if gotPath != "/api/v1/network/assets" {
		t.Errorf("path = %q, want /api/v1/network/assets", gotPath)
	}
	if !strings.Contains(gotQuery, "size=20") {
		t.Errorf("query %q missing size=20 (should use 'size' not 'limit')", gotQuery)
	}
	if !strings.Contains(gotQuery, "page=0") {
		t.Errorf("query %q missing page=0", gotQuery)
	}
	if !strings.Contains(gotQuery, "name=Office") {
		t.Errorf("query %q missing name=Office", gotQuery)
	}
	if !strings.Contains(gotQuery, "category=router") {
		t.Errorf("query %q missing category=router", gotQuery)
	}
	if !strings.Contains(gotQuery, "status=in_use") {
		t.Errorf("query %q missing status=in_use", gotQuery)
	}
}

func TestAssetCreate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": sampleAsset,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "create",
		"--name", "Office Router",
		"--mac", "00:18:05:AB:CD:EF",
		"--category", "router",
		"--status", "in_use",
		"--number", "AST-001",
		"--warranty-expiration", "2027-12-31",
		"--notes", "2nd floor",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset create: %v", err)
	}

	if gotPath != "/api/v1/network/assets" {
		t.Errorf("path = %q, want /api/v1/network/assets", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["name"] != "Office Router" {
		t.Errorf("body.name = %v, want Office Router", gotBody["name"])
	}
	if gotBody["mac"] != "00:18:05:AB:CD:EF" {
		t.Errorf("body.mac = %v, want 00:18:05:AB:CD:EF", gotBody["mac"])
	}
	if gotBody["category"] != "router" {
		t.Errorf("body.category = %v, want router", gotBody["category"])
	}
}

func TestAssetUpdate(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": sampleAsset,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "update", "67c8d537e7f8d2c1e5fffdaa",
		"--name", "Updated Router",
		"--category", "router",
		"--status", "in_repair",
		"--notes", "needs fixing",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset update: %v", err)
	}

	if gotPath != "/api/v1/network/assets/67c8d537e7f8d2c1e5fffdaa" {
		t.Errorf("path = %q, want /api/v1/network/assets/67c8d537e7f8d2c1e5fffdaa", gotPath)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotBody["name"] != "Updated Router" {
		t.Errorf("body.name = %v, want Updated Router", gotBody["name"])
	}
	if gotBody["status"] != "in_repair" {
		t.Errorf("body.status = %v, want in_repair", gotBody["status"])
	}
}

func TestAssetUpdatePartial(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": sampleAsset,
		})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "update", "67c8d537e7f8d2c1e5fffdaa",
		"--notes", "updated note",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset update partial: %v", err)
	}

	if _, exists := gotBody["name"]; exists {
		t.Error("body should not contain 'name' when flag not provided")
	}
	if _, exists := gotBody["category"]; exists {
		t.Error("body should not contain 'category' when flag not provided")
	}
	if gotBody["notes"] != "updated note" {
		t.Errorf("body.notes = %v, want 'updated note'", gotBody["notes"])
	}
}

func TestAssetUpdateNoFlags(t *testing.T) {
	f, _ := newTestFactory(t, "http://unused")
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "update", "67c8d537e7f8d2c1e5fffdaa"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
	if err.Error() != "at least one field must be specified" {
		t.Errorf("error = %q, want 'at least one field must be specified'", err.Error())
	}
}

func TestAssetDeleteSingle(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		json.NewEncoder(w).Encode(map[string]interface{}{"result": sampleAsset})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "delete", "67c8d537e7f8d2c1e5fffdaa", "-y"})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset delete: %v", err)
	}

	if gotPath != "/api/v1/network/assets/67c8d537e7f8d2c1e5fffdaa" {
		t.Errorf("path = %q, want /api/v1/network/assets/67c8d537e7f8d2c1e5fffdaa", gotPath)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestAssetDeleteBatch(t *testing.T) {
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
	root := newAssetRoot(f)
	root.SetArgs([]string{"asset", "delete", "id1", "id2", "id3", "-y"})
	if err := root.Execute(); err != nil {
		t.Fatalf("asset delete batch: %v", err)
	}

	if gotPath != "/api/v1/network/assets/remove" {
		t.Errorf("path = %q, want /api/v1/network/assets/remove", gotPath)
	}
	if gotMethod != "POST" {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	ids, ok := gotBody["ids"].([]interface{})
	if !ok || len(ids) != 3 {
		t.Errorf("body.ids = %v, want [id1 id2 id3]", gotBody["ids"])
	}
}
