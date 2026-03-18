package iostreams

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ColumnFormatter transforms a display string for a specific column.
type ColumnFormatter func(string) string

// ColumnFormatters maps column names (gjson paths) to their formatter functions.
type ColumnFormatters map[string]ColumnFormatter

// FormatBytes converts a numeric string (bytes) to a human-readable size.
//
//	"21114126336" → "19.7 GB"
//	"1024"        → "1.0 KB"
//	"not-a-number"→ "not-a-number" (returned as-is)
func FormatBytes(s string) string {
	b, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	const (
		KB = 1024.0
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case b >= TB:
		return fmt.Sprintf("%.1f TB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.1f GB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", b/KB)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// FormatBitRate converts a numeric string (bits per second) to a human-readable rate.
//
//	"1000000"    → "1.0 Mbps"
//	"1500"       → "1.5 Kbps"
//	"not-a-number"→ "not-a-number" (returned as-is)
func FormatBitRate(s string) string {
	b, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return s
	}
	const (
		Kbps = 1000.0
		Mbps = Kbps * 1000
		Gbps = Mbps * 1000
	)
	switch {
	case b >= Gbps:
		return fmt.Sprintf("%.1f Gbps", b/Gbps)
	case b >= Mbps:
		return fmt.Sprintf("%.1f Mbps", b/Mbps)
	case b >= Kbps:
		return fmt.Sprintf("%.1f Kbps", b/Kbps)
	default:
		return fmt.Sprintf("%.0f bps", b)
	}
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

	d := time.Since(t)
	if d < 0 {
		d = -d
		return formatTimeDistance(d) + " from now"
	}
	return formatTimeDistance(d) + " ago"
}

func formatTimeDistance(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	case d < 30*24*time.Hour:
		days := int(d.Hours()) / 24
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	default:
		months := int(d.Hours()) / (24 * 30)
		if months == 0 {
			months = 1
		}
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}
}
