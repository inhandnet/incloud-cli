package device

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestConfigCopy_RejectsUnsupportedCombinations(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "group to devices",
			args:    []string{"--source-group", "GRP1", "--target", "DEV1", "--yes"},
			wantErr: "copying a group config to devices is not supported",
		},
		{
			name:    "device to devices and groups",
			args:    []string{"--source", "DEV1", "--target", "DEV2", "--target-group", "GRP1", "--yes"},
			wantErr: "--target and --target-group are mutually exclusive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))
			defer server.Close()

			f, _ := newTestFactory(t, server.URL)
			cmd := newCmdConfigCopy(f)
			cmd.SetArgs(tt.args)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
			if called {
				t.Error("expected no HTTP request")
			}
		})
	}
}

func TestConfigCopy_SupportedCombinations(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantBody map[string]any
	}{
		{
			name:     "device to devices",
			args:     []string{"--source", "DEV1", "--target", "DEV2", "--target", "DEV3"},
			wantBody: map[string]any{"sourceDeviceId": "DEV1", "targetDeviceIds": []any{"DEV2", "DEV3"}},
		},
		{
			name:     "device to groups",
			args:     []string{"--source", "DEV1", "--target-group", "GRP1"},
			wantBody: map[string]any{"sourceDeviceId": "DEV1", "targetGroupIds": []any{"GRP1"}},
		},
		{
			name:     "group to groups",
			args:     []string{"--source-group", "GRP1", "--target-group", "GRP2", "--module", "wlan_ap"},
			wantBody: map[string]any{"sourceGroupId": "GRP1", "targetGroupIds": []any{"GRP2"}, "module": "wlan_ap"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				_, _ = w.Write([]byte(`{"result":"ok"}`))
			}))
			defer server.Close()

			f, _ := newTestFactory(t, server.URL)
			cmd := newCmdConfigCopy(f)
			cmd.SetArgs(append(tt.args, "--yes"))
			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != "/api/v1/config/layer/bulk-copy" {
				t.Errorf("unexpected path: %s", gotPath)
			}
			if !reflect.DeepEqual(gotBody, tt.wantBody) {
				t.Errorf("unexpected body: %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
