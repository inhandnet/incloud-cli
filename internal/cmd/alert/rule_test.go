package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// backendRuleTypes mirrors nezha-alert rule/RuleType.java (lowercased). When the
// backend adds a type, add it here and to ruleTypeRegistry.
var backendRuleTypes = []string{
	"connected", "disconnected", "config_sync_failed", "sim_switch", "local_config_update",
	"reboot", "firmware_upgrade", "license_expiring", "license_expired", "uplink_switch",
	"ethernet_wan_connected", "ethernet_wan_disconnected", "modem_wan_connected", "modem_wan_disconnected",
	"wwan_connected", "wwan_disconnected", "client_connected", "client_disconnected",
	"cell_operator_switch", "bridge_loop_detect", "cell_traffic_reach_threshold", "device_power_off",
	"uplink_status_change", "high_average_cpu_utilization", "high_memory_utilization",
	"poor_cellular_signal_strength", "posready_ineffective", "device_login_locked",
	"ap_offline", "ap_config_sync_fail", "ap_config_sync_ok", "roam_fail", "roam_success",
}

func TestRuleTypeRegistry_MatchesBackend(t *testing.T) {
	var got []string
	for _, def := range AllRuleTypes() {
		got = append(got, def.Type)
	}
	want := append([]string(nil), backendRuleTypes...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("registry types differ from backend RuleType\n got: %v\nwant: %v", got, want)
	}
}

// existingRule is a GET /alerts/rules/{id} response whose notify carries fields
// the CLI has no flags for (emails) and response-only details.
var existingRule = map[string]any{
	"result": map[string]any{
		"_id":       "507f1f77bcf86cd799439011",
		"type":      "org",
		"targetIds": []string{"607f1f77bcf86cd799439022"},
		"notify": map[string]any{
			"channels":         []string{"EMAIL"},
			"users":            []string{"707f1f77bcf86cd799439033"},
			"emails":           []string{"ops@example.com"},
			"activeDayOfWeeks": []string{"MONDAY", "FRIDAY"},
			"startTime":        "09:00",
			"endTime":          "18:00",
			"timezone":         "Asia/Shanghai",
			"userDetails":      []any{map[string]any{"name": "x"}},
		},
	},
}

func runRuleUpdate(t *testing.T, args ...string) (map[string]any, error) {
	t.Helper()
	var putBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(existingRule)
			return
		}
		json.NewDecoder(r.Body).Decode(&putBody)
		json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{}})
	}))
	defer srv.Close()

	f, _ := newTestFactory(t, srv.URL)
	root := newAlertRoot(f)
	root.SetArgs(append([]string{"alert", "rule", "update", "507f1f77bcf86cd799439011", "--type", "reboot"}, args...))
	err := root.Execute()
	if putBody == nil {
		return nil, err
	}
	notify, _ := putBody["notify"].(map[string]any)
	return notify, err
}

func TestRuleUpdate_PreservesNotifyWhenNoNotifyFlags(t *testing.T) {
	notify, err := runRuleUpdate(t)
	if err != nil {
		t.Fatalf("rule update: %v", err)
	}
	for _, k := range []string{"channels", "users", "emails", "activeDayOfWeeks", "startTime", "endTime", "timezone"} {
		if _, ok := notify[k]; !ok {
			t.Errorf("notify.%s dropped: %v", k, notify)
		}
	}
	if notify["timezone"] != "Asia/Shanghai" {
		t.Errorf("timezone = %v, want Asia/Shanghai", notify["timezone"])
	}
	if _, ok := notify["userDetails"]; ok {
		t.Errorf("response-only userDetails must not be sent back")
	}
}

func TestRuleUpdate_OverridesOnlyChangedFlags(t *testing.T) {
	notify, err := runRuleUpdate(t, "--start-time", "08:00", "--timezone", "Europe/Berlin")
	if err != nil {
		t.Fatalf("rule update: %v", err)
	}
	if notify["startTime"] != "08:00" || notify["endTime"] != "18:00" {
		t.Errorf("time window = %v-%v, want 08:00-18:00", notify["startTime"], notify["endTime"])
	}
	if notify["timezone"] != "Europe/Berlin" {
		t.Errorf("timezone = %v, want Europe/Berlin", notify["timezone"])
	}
	if emails, _ := notify["emails"].([]any); len(emails) != 1 {
		t.Errorf("emails = %v, want preserved", notify["emails"])
	}
}

func TestRuleUpdate_EmptyValueClears(t *testing.T) {
	notify, err := runRuleUpdate(t, "--timezone", "", "--user", "")
	if err != nil {
		t.Fatalf("rule update: %v", err)
	}
	if _, ok := notify["timezone"]; ok {
		t.Errorf("timezone should be cleared: %v", notify)
	}
	if _, ok := notify["users"]; ok {
		t.Errorf("users should be cleared: %v", notify)
	}
}

func TestRuleUpdate_DropsEmptyListItems(t *testing.T) {
	notify, err := runRuleUpdate(t, "--user", "", "--user", "807f1f77bcf86cd799439044")
	if err != nil {
		t.Fatalf("rule update: %v", err)
	}
	if users, _ := notify["users"].([]any); len(users) != 1 || users[0] != "807f1f77bcf86cd799439044" {
		t.Errorf("users = %v, want only the non-empty id", notify["users"])
	}
}

func TestRuleUpdate_RejectsClearingChannels(t *testing.T) {
	_, err := runRuleUpdate(t, "--channel", "")
	if err == nil || !strings.Contains(err.Error(), "--channel") {
		t.Fatalf("err = %v, want missing channel error", err)
	}
}

func TestNotifyValidate_Timezone(t *testing.T) {
	for tz, ok := range map[string]bool{"": true, "Asia/Shanghai": true, "UTC": true, "Mars/Base": false, "Local": false} {
		err := (&NotifyOptions{Timezone: tz}).Validate()
		if (err == nil) != ok {
			t.Errorf("Validate(%q) err = %v, want ok=%v", tz, err, ok)
		}
	}
}
