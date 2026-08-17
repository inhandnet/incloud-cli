package unitdecl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inhandnet/incloud-cli/internal/iostreams"
)

// TestNoSuspectedSentinelSet guards IM-3180 acceptance criterion 7: no "suspected
// timeout clamp" value set may exist anywhere in the tree. 2000000 microseconds
// is a real measurement on a degraded link (the only sentinel is -1, per
// nezha-agent/pkg/message/message.go:117), so annotating it would mask a genuine
// fault.
//
// The scan covers the whole repository rather than one file: the rules used to
// live in internal/iostreams and now live here, and the next move should not
// silently disarm this guard.
func TestNoSuspectedSentinelSet(t *testing.T) {
	banned := []string{"suspected", "Suspected", "no-measurement", "NoMeasurement", "suspectedTimeout"}
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); name == ".git" || name == "bin" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		// Test files legitimately name the banned strings while asserting their
		// absence; documentation may describe the corrected history.
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, b := range banned {
			if strings.Contains(string(src), b) {
				t.Errorf("%s mentions %q; the suspected-sentinel set must stay deleted", path, b)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestDeclarationsCarryUnits checks that every declared rewrite actually renames
// the field to a name ending in a unit word. A declaration whose target name
// does not name a unit would defeat the entire purpose of this package.
func TestDeclarationsCarryUnits(t *testing.T) {
	all := map[string]iostreams.FieldRewrites{
		"DeviceUplink":     DeviceUplink,
		"DeviceUplinkGet":  DeviceUplinkGet,
		"DeviceUplinkPerf": DeviceUplinkPerf,
		"DeviceInterface":  DeviceInterface,
		"DeviceLogMqtt":    DeviceLogMqtt,
		"OverviewOffline":  OverviewOffline,
	}
	units := []string{"Microseconds", "Milliseconds", "Seconds", "Bytes", "Percent"}
	for name, rules := range all {
		if len(rules) == 0 {
			t.Errorf("%s declares no rewrites", name)
		}
		for from, rule := range rules {
			hasUnit := false
			for _, u := range units {
				if strings.HasSuffix(rule.To, u) {
					hasUnit = true
				}
			}
			if !hasUnit {
				t.Errorf("%s: %q renames to %q, which names no unit", name, from, rule.To)
			}
			if rule.Timeout && rule.StatusKey == "" {
				t.Errorf("%s: %q annotates the timeout sentinel but has no StatusKey", name, from)
			}
		}
	}
}
