package cmdutil

import (
	"fmt"
	"time"
	_ "time/tzdata" // embed IANA zone data so --timezone validation works on Windows
)

// ValidateTimezone checks that a --timezone value is an IANA zone name the
// server understands. Empty is accepted (the server falls back to its default).
func ValidateTimezone(tz string) error {
	if tz == "" {
		return nil
	}
	// time.LoadLocation accepts "Local", which is not an IANA name the server understands.
	if tz == "Local" {
		return fmt.Errorf("invalid --timezone %q: must be an IANA timezone name such as Asia/Shanghai", tz)
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fmt.Errorf("invalid --timezone %q: must be an IANA timezone name such as Asia/Shanghai", tz)
	}
	return nil
}
