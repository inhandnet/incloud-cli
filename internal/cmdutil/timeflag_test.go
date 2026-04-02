package cmdutil

import (
	"testing"
	"time"
)

func TestParseTimeFlag(t *testing.T) {
	// Fix local timezone to Asia/Shanghai (+08:00) for deterministic tests.
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("failed to load Asia/Shanghai: %v", err)
	}
	origLocal := time.Local
	time.Local = loc
	t.Cleanup(func() { time.Local = origLocal })

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"utc Z suffix", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z"},
		{"with positive offset", "2025-01-01T08:00:00+08:00", "2025-01-01T00:00:00Z"},
		{"with negative offset", "2025-01-01T00:00:00-05:00", "2025-01-01T05:00:00Z"},
		{"iso without tz", "2025-01-01T08:00:00", "2025-01-01T00:00:00Z"},
		{"space separator", "2025-01-01 08:00:00", "2025-01-01T00:00:00Z"},
		{"date only", "2025-01-01", "2024-12-31T16:00:00Z"},
		{"date only midnight", "2025-07-01", "2025-06-30T16:00:00Z"},
		{"invalid", "not-a-date", "not-a-date"},
		{"partial", "2025-01", "2025-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTimeFlag(tt.input)
			if got != tt.want {
				t.Errorf("ParseTimeFlag(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
