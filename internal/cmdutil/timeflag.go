package cmdutil

import "time"

// localTimeLayouts are formats without timezone info, parsed as local time.
var localTimeLayouts = []string{
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// ParseTimeFlag normalises a user-supplied time string to UTC RFC 3339.
//
// Accepted inputs:
//   - RFC 3339 with timezone (2025-01-01T00:00:00Z, 2025-01-01T08:00:00+08:00)
//   - ISO 8601 without timezone (2025-01-01T08:00:00) → treated as local time
//   - Date only (2025-01-01) → 00:00:00 local time
//   - Empty string → returned as-is
//   - Anything else → returned as-is (let the API report the error)
func ParseTimeFlag(s string) string {
	if s == "" {
		return ""
	}

	// Already has timezone info (Z or +/-offset).
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format(time.RFC3339)
	}

	// No timezone — interpret as local time.
	for _, layout := range localTimeLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.UTC().Format(time.RFC3339)
		}
	}

	return s
}
