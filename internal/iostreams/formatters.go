package iostreams

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

// ColumnFormatter transforms a display string for a specific column.
type ColumnFormatter func(string) string

// ColumnFormatters maps column names (gjson paths) to their formatter functions.
type ColumnFormatters map[string]ColumnFormatter

// FormatBytes converts a numeric string (bytes) to a human-readable size (IEC, 1024-based).
//
//	"21114126336" → "20 GiB"
//	"1024"        → "1.0 KiB"
//	"not-a-number"→ "not-a-number" (returned as-is)
func FormatBytes(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return humanize.IBytes(uint64(f))
}

// FormatBitRate converts a numeric string (bits per second) to a human-readable rate.
//
//	"1000000"    → "1 Mbps"
//	"1500"       → "1.5 kbps"
//	"not-a-number"→ "not-a-number" (returned as-is)
func FormatBitRate(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return humanize.SIWithDigits(f, 1, "bps")
}

// FormatMbps appends " Mbps" to a numeric string.
//
//	"25.5" → "25.50 Mbps"
//	"0"    → "0.00 Mbps"
func FormatMbps(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.2f Mbps", f)
}

// FormatMicroseconds converts a numeric string (microseconds) to a human-readable latency.
//
//	"4765"  → "4.765 ms"
//	"1200000" → "1200.000 ms"
//	"500"   → "0.500 ms"
func FormatMicroseconds(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.3f ms", f/1000)
}

// FormatMs appends " ms" to a numeric string.
//
//	"12.5" → "12.50 ms"
//	"0"    → "0.00 ms"
func FormatMs(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.2f ms", f)
}

// FormatPercent converts a decimal fraction string to a percentage.
//
//	"0.452"  → "45.2%"
//	"1.0"    → "100.0%"
func FormatPercent(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%.1f%%", f*100)
}

// FormatDuration converts a numeric string (seconds) to a human-readable duration.
//
//	"3661"  → "1h 1m 1s"
//	"45"    → "45s"
//	"90061" → "1d 1h 1m"
func FormatDuration(s string) string {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	d := time.Duration(f * float64(time.Second))
	if d == 0 {
		return "0s"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if mins > 0 {
		parts = append(parts, fmt.Sprintf("%dm", mins))
	}
	// Only show seconds if total < 1 hour
	if secs > 0 && days == 0 && hours == 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}
	if len(parts) == 0 {
		return "0s"
	}
	return strings.Join(parts, " ")
}

// FormatRelativeTime converts an ISO 8601 timestamp to a relative time string.
//
//	"2026-03-17T10:00:00Z" → "2 hours ago" (if now is 12:00)
func FormatRelativeTime(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	var t time.Time
	var err error
	for _, layout := range formats {
		t, err = time.Parse(layout, s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return s
	}

	return humanize.Time(t)
}

// TruncateRunes truncates s to maxLen runes and appends "..." if it was longer.
func TruncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
