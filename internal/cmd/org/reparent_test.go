package org

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newReparentServer(t *testing.T, status int, resp map[string]any) (*httptest.Server, *string, map[string]any) {
	t.Helper()
	var gotPath string
	gotBody := map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotPath, gotBody
}

func TestOrgReparent_PostsExpectedBody(t *testing.T) {
	srv, gotPath, gotBody := newReparentServer(t, http.StatusOK, map[string]any{
		"result": map[string]any{"_id": "61259f8f4be3e571fcfa4d75", "name": "Acme"},
	})
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "reparent",
		"--moving", "61259f8f4be3e571fcfa4d75",
		"--new-parent", "6125a0114be3e571fcfa4d80",
		"--move-billing-assets",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("org reparent: %v", err)
	}

	if *gotPath != "/api/v1/orgs/reparent" {
		t.Errorf("path = %q, want /api/v1/orgs/reparent", *gotPath)
	}
	if gotBody["movingOrgId"] != "61259f8f4be3e571fcfa4d75" {
		t.Errorf("movingOrgId = %v", gotBody["movingOrgId"])
	}
	if gotBody["newParentId"] != "6125a0114be3e571fcfa4d80" {
		t.Errorf("newParentId = %v", gotBody["newParentId"])
	}
	if gotBody["moveBillingAssets"] != true {
		t.Errorf("moveBillingAssets = %v, want true", gotBody["moveBillingAssets"])
	}
}

func TestOrgReparent_DefaultMoveBillingAssetsFalse(t *testing.T) {
	srv, _, gotBody := newReparentServer(t, http.StatusOK, map[string]any{
		"result": map[string]any{"_id": "x", "name": "y"},
	})
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "reparent", "--moving", "a", "--new-parent", "b"})
	if err := root.Execute(); err != nil {
		t.Fatalf("org reparent: %v", err)
	}
	if gotBody["moveBillingAssets"] != false {
		t.Errorf("moveBillingAssets = %v, want false", gotBody["moveBillingAssets"])
	}
}

func TestOrgReparent_PrecheckErrorSurfaced(t *testing.T) {
	srv, _, _ := newReparentServer(t, http.StatusBadRequest, map[string]any{
		"error": map[string]any{"code": "INVALID_PARAMETER", "message": "reparent would create a cycle"},
	})
	f := newTestFactory(t, srv.URL)
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "reparent", "--moving", "a", "--new-parent", "b"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error on precheck failure, got nil")
	}
	out := f.IO.Out.(*bytes.Buffer).String()
	if !strings.Contains(out, "cycle") {
		t.Errorf("precheck message not surfaced to user; output = %q", out)
	}
}

func TestOrgReparent_RequiredFlags(t *testing.T) {
	f := newTestFactory(t, "http://127.0.0.1:0")
	root := newOrgRoot(f)
	root.SetArgs([]string{"org", "reparent", "--moving", "a"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when --new-parent missing, got nil")
	}
}
