package cmdutil

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNoArgsWithHint(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}
	args := NoArgsWithHint(DeviceFilterHint("incloud touch client list", "device-id", "incloud device list -q %s"))

	if err := args(cmd, nil); err != nil {
		t.Fatalf("no args: unexpected error %v", err)
	}

	tests := []struct {
		arg  string
		want string
	}{
		{"507f1f77bcf86cd799439011", "run 'incloud touch client list --device-id 507f1f77bcf86cd799439011'"},
		{"SN123", "run 'incloud device list -q SN123' to find its id"},
	}
	for _, tt := range tests {
		err := args(cmd, []string{tt.arg})
		if err == nil {
			t.Fatalf("%s: expected error", tt.arg)
		}
		msg := err.Error()
		if !strings.Contains(msg, `unexpected argument "`+tt.arg+`"`) || !strings.Contains(msg, tt.want) {
			t.Errorf("%s: got %q, want it to contain %q", tt.arg, msg, tt.want)
		}
	}
}
