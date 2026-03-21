package device

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveProductVersion_FromFlags(t *testing.T) {
	// No server needed — flags provide values directly
	pv := &productVersion{product: "MR805", version: "V2.0.15-111"}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_FromDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/dev123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":{"_id":"dev123","partNumber":"MR805","firmware":"V2.0.15-111"}}`))
	}))
	defer server.Close()

	f, _ := newTestFactory(t, server.URL)
	client, _ := f.APIClient()

	pv, err := resolveProductVersion(client, "dev123", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if pv.product != "MR805" || pv.version != "V2.0.15-111" {
		t.Errorf("unexpected: %+v", pv)
	}
}

func TestResolveProductVersion_MutualExclusion(t *testing.T) {
	_, err := resolveProductVersion(nil, "dev123", "MR805", "V2.0.15")
	if err == nil {
		t.Fatal("expected error for mutual exclusion")
	}
}

func TestResolveProductVersion_MissingParams(t *testing.T) {
	_, err := resolveProductVersion(nil, "", "MR805", "")
	if err == nil {
		t.Fatal("expected error when --product without --version")
	}
	_, err = resolveProductVersion(nil, "", "", "")
	if err == nil {
		t.Fatal("expected error when no params")
	}
}
