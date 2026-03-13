package api

import (
	"net/http"
	"testing"
)

func TestBuildRequest_GET(t *testing.T) {
	opts := &ApiOptions{
		Path:        "/api/v1/devices",
		Method:      "GET",
		QueryParams: []string{"page=0", "limit=10"},
		Host:        "https://portal.example.com",
	}
	req, err := buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != http.MethodGet {
		t.Errorf("expected GET, got %s", req.Method)
	}
	if req.URL.String() != "https://portal.example.com/api/v1/devices?limit=10&page=0" {
		t.Errorf("unexpected URL: %s", req.URL.String())
	}
}

func TestBuildRequest_POST_Fields(t *testing.T) {
	opts := &ApiOptions{
		Path:       "/api/v1/devices",
		Method:     "POST",
		BodyFields: []string{"name=test", "product=router"},
		Host:       "https://portal.example.com",
	}
	req, err := buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != http.MethodPost {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected JSON content type")
	}
}

func TestBuildRequest_Headers(t *testing.T) {
	opts := &ApiOptions{
		Path:    "/api/v1/users/me",
		Method:  "GET",
		Headers: []string{"Sudo: admin@example.com", "X-Custom: value"},
		Host:    "https://portal.example.com",
	}
	req, err := buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Sudo") != "admin@example.com" {
		t.Errorf("missing Sudo header")
	}
}

func TestBuildRequest_DefaultMethod(t *testing.T) {
	opts := &ApiOptions{
		Path: "/api/v1/test",
		Host: "https://portal.example.com",
	}
	req, err := buildRequest(opts)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != http.MethodGet {
		t.Errorf("expected default GET, got %s", req.Method)
	}
}
