package api

import (
	"testing"
)

func TestBuildRequestOptions_GET(t *testing.T) {
	opts := &ApiOptions{
		Path:        "/api/v1/devices",
		Method:      "GET",
		QueryParams: []string{"page=0", "limit=10"},
	}
	reqOpts, err := buildRequestOptions(opts)
	if err != nil {
		t.Fatal(err)
	}
	if reqOpts.Query.Get("page") != "0" {
		t.Errorf("expected page=0, got %s", reqOpts.Query.Get("page"))
	}
	if reqOpts.Query.Get("limit") != "10" {
		t.Errorf("expected limit=10, got %s", reqOpts.Query.Get("limit"))
	}
}

func TestBuildRequestOptions_POST_Fields(t *testing.T) {
	opts := &ApiOptions{
		Path:       "/api/v1/devices",
		Method:     "POST",
		BodyFields: []string{"name=test", "product=router"},
	}
	reqOpts, err := buildRequestOptions(opts)
	if err != nil {
		t.Fatal(err)
	}
	body, ok := reqOpts.Body.(map[string]interface{})
	if !ok {
		t.Fatal("expected Body to be map[string]interface{}")
	}
	if body["name"] != "test" {
		t.Errorf("expected name=test, got %v", body["name"])
	}
	if body["product"] != "router" {
		t.Errorf("expected product=router, got %v", body["product"])
	}
}

func TestBuildRequestOptions_Headers(t *testing.T) {
	opts := &ApiOptions{
		Path:    "/api/v1/users/me",
		Method:  "GET",
		Headers: []string{"Sudo: admin@example.com", "X-Custom: value"},
	}
	reqOpts, err := buildRequestOptions(opts)
	if err != nil {
		t.Fatal(err)
	}
	if reqOpts.Headers["Sudo"] != "admin@example.com" {
		t.Errorf("missing Sudo header")
	}
	if reqOpts.Headers["X-Custom"] != "value" {
		t.Errorf("missing X-Custom header")
	}
}

func TestBuildRequestOptions_DefaultMethod(t *testing.T) {
	opts := &ApiOptions{
		Path: "/api/v1/test",
	}
	reqOpts, err := buildRequestOptions(opts)
	if err != nil {
		t.Fatal(err)
	}
	// No body, no query, no headers — all nil/empty
	if reqOpts.Query != nil {
		t.Errorf("expected nil query, got %v", reqOpts.Query)
	}
	if reqOpts.Body != nil {
		t.Errorf("expected nil body")
	}
}
