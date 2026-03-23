package alert

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RuleParam defines a parameter for an alert rule type.
type RuleParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`        // "integer", "number", "string"
	Unit        string `json:"unit"`        // "seconds", "percent", "dBm", "dB"
	Description string `json:"description"` // human-readable description
}

// RuleTypeDef defines an alert rule type and its accepted parameters.
type RuleTypeDef struct {
	Type        string      `json:"type"`
	Category    string      `json:"category"` // "general" or "network"
	Description string      `json:"description"`
	Params      []RuleParam `json:"params,omitempty"`
}

// ruleTypeRegistry is the canonical list of supported alert rule types.
var ruleTypeRegistry = []RuleTypeDef{
	// General — device state
	{Type: "connected", Category: "general", Description: "Device connects to platform", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering (e.g. 600 = 10 minutes, range 60-1800)"},
	}},
	{Type: "disconnected", Category: "general", Description: "Device disconnects from platform", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering (e.g. 600 = 10 minutes, range 60-1800)"},
	}},
	{Type: "reboot", Category: "general", Description: "Device reboots"},
	{Type: "firmware_upgrade", Category: "general", Description: "Firmware changes"},
	{Type: "device_power_off", Category: "general", Description: "Device powered down"},

	// General — config
	{Type: "config_sync_failed", Category: "general", Description: "Configuration sync failure"},
	{Type: "local_config_update", Category: "general", Description: "Local configuration changed"},

	// General — license
	{Type: "license_expiring", Category: "general", Description: "License expiring soon", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Advance notice in seconds (e.g. 2592000 = 30 days, range 86400-2592000)"},
	}},
	{Type: "license_expired", Category: "general", Description: "License expired"},

	// General — client
	{Type: "client_connected", Category: "general", Description: "Client comes online", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering (e.g. 600 = 10 minutes, range 60-1800)"},
	}},
	{Type: "client_disconnected", Category: "general", Description: "Client goes offline", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering (e.g. 600 = 10 minutes, range 60-1800)"},
	}},

	// General — resource
	{Type: "high_average_cpu_utilization", Category: "general", Description: "CPU usage stays above threshold", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Monitoring window in seconds (300/600/900/1800)"},
		{Name: "threshold", Type: "integer", Unit: "percent", Description: "CPU usage threshold (70, 80, or 90)"},
	}},
	{Type: "high_memory_utilization", Category: "general", Description: "Memory usage stays above threshold", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Monitoring window in seconds (300/600/900/1800)"},
		{Name: "threshold", Type: "integer", Unit: "percent", Description: "Memory usage threshold (70, 80, or 90)"},
	}},
	{Type: "cell_traffic_reach_threshold", Category: "general", Description: "Cellular data usage reaches threshold"},

	// General — cellular signal
	{Type: "poor_cellular_signal_strength", Category: "general", Description: "Cellular signal below threshold", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering"},
		{Name: "rsrpThreshold", Type: "number", Unit: "dBm", Description: "RSRP threshold in dBm (e.g. -116)"},
		{Name: "sinrThreshold", Type: "number", Unit: "dB", Description: "SINR threshold in dB (e.g. 0)"},
	}},

	// Network
	{Type: "sim_switch", Category: "network", Description: "SIM card switched"},
	{Type: "uplink_switch", Category: "network", Description: "Primary uplink switched"},
	{Type: "ethernet_wan_connected", Category: "network", Description: "Ethernet WAN connected"},
	{Type: "ethernet_wan_disconnected", Category: "network", Description: "Ethernet WAN disconnected"},
	{Type: "modem_wan_connected", Category: "network", Description: "Cellular WAN connected"},
	{Type: "modem_wan_disconnected", Category: "network", Description: "Cellular WAN disconnected"},
	{Type: "wwan_connected", Category: "network", Description: "Wi-Fi(STA) WAN connected"},
	{Type: "wwan_disconnected", Category: "network", Description: "Wi-Fi(STA) WAN disconnected"},
	{Type: "bridge_loop_detect", Category: "network", Description: "Bridge loop detected"},
	{Type: "cell_operator_switch", Category: "network", Description: "Cellular carrier switched"},
	{Type: "uplink_status_change", Category: "network", Description: "Uplink interface status change", Params: []RuleParam{
		{Name: "retention", Type: "integer", Unit: "seconds", Description: "Duration in seconds before triggering (60/300/900/1800)"},
		{Name: "interface", Type: "string", Unit: "", Description: "Interface name (primary_link, wan1-wan4, cellular1-cellular4, wi-fi(sta))"},
	}},
}

// ruleTypeIndex provides O(1) lookup by type name (lowercase).
var ruleTypeIndex map[string]*RuleTypeDef

func init() {
	ruleTypeIndex = make(map[string]*RuleTypeDef, len(ruleTypeRegistry))
	for i := range ruleTypeRegistry {
		ruleTypeIndex[ruleTypeRegistry[i].Type] = &ruleTypeRegistry[i]
	}
}

// LookupRuleType returns the definition for a given type name (case-insensitive).
func LookupRuleType(name string) (*RuleTypeDef, bool) {
	def, ok := ruleTypeIndex[strings.ToLower(name)]
	return def, ok
}

// AllRuleTypes returns the full registry.
func AllRuleTypes() []RuleTypeDef {
	return ruleTypeRegistry
}

// ParsedRule represents a parsed --type flag value.
type ParsedRule struct {
	Type  string         `json:"type"`
	Param map[string]any `json:"param,omitempty"`
}

// ParseTypeFlag parses a single --type flag value.
// Accepted formats:
//   - "reboot"                                          (type only)
//   - "disconnected,retention=600"                      (comma-separated key=value)
//   - '{"type":"disconnected","param":{"retention":600}}' (JSON)
func ParseTypeFlag(value string) (ParsedRule, error) {
	value = strings.TrimSpace(value)

	// JSON format
	if strings.HasPrefix(value, "{") {
		var rule ParsedRule
		if err := json.Unmarshal([]byte(value), &rule); err != nil {
			return ParsedRule{}, fmt.Errorf("invalid JSON rule: %w", err)
		}
		if rule.Type == "" {
			return ParsedRule{}, fmt.Errorf("JSON rule missing \"type\" field")
		}
		rule.Type = strings.ToLower(rule.Type)
		return rule, nil
	}

	// Comma-separated format: type[,key=value,...]
	parts := strings.Split(value, ",")
	typeName := strings.ToLower(strings.TrimSpace(parts[0]))
	if typeName == "" {
		return ParsedRule{}, fmt.Errorf("empty type name")
	}

	rule := ParsedRule{Type: typeName}
	if len(parts) > 1 {
		rule.Param = make(map[string]any, len(parts)-1)
		for _, kv := range parts[1:] {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return ParsedRule{}, fmt.Errorf("invalid param %q: expected key=value", kv)
			}
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			if k == "" {
				return ParsedRule{}, fmt.Errorf("invalid param %q: key must not be empty", kv)
			}
			// Try to parse as number
			rule.Param[k] = parseParamValue(v)
		}
	}

	return rule, nil
}

// ParseTypeFlags parses all --type flag values into rules.
func ParseTypeFlags(values []string) ([]ParsedRule, error) {
	rules := make([]ParsedRule, 0, len(values))
	for _, v := range values {
		rule, err := ParseTypeFlag(v)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// RulesToRequestBody converts parsed rules to the API request body format.
func RulesToRequestBody(rules []ParsedRule) []map[string]any {
	result := make([]map[string]any, len(rules))
	for i, r := range rules {
		entry := map[string]any{"type": r.Type}
		if r.Param != nil {
			entry["param"] = r.Param
		} else {
			entry["param"] = map[string]any{}
		}
		result[i] = entry
	}
	return result
}

// parseParamValue tries to parse a string as a number, falls back to string.
func parseParamValue(s string) any {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}
