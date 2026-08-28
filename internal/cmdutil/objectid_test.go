package cmdutil

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestIsObjectID(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{"valid lowercase hex", "507f1f77bcf86cd799439011", true},
		{"valid uppercase hex", "507F1F77BCF86CD799439011", true},
		{"valid mixed case hex", "507f1F77Bcf86CD799439011", true},
		{"too short", "17572", false},
		{"too long", "507f1f77bcf86cd799439011ab", false},
		{"non-hex characters", "507f1f77bcf86cd79943901g", false},
		{"empty string", "", false},
		{"serial number style", "SN-1234-ABCD", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsObjectID(tc.value); got != tc.want {
				t.Errorf("IsObjectID(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}

func requireErrContains(t *testing.T, err error, substrs ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	msg := err.Error()
	for _, s := range substrs {
		if !strings.Contains(msg, s) {
			t.Errorf("error %q does not contain %q", msg, s)
		}
	}
}

func TestObjectIDArgs(t *testing.T) {
	validator := ObjectIDArgs(cobra.ExactArgs(1), 0, "device id", "incloud device list -q %s")
	cmd := &cobra.Command{}

	t.Run("valid id passes", func(t *testing.T) {
		if err := validator(cmd, []string{"507f1f77bcf86cd799439011"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid id reports original value, example format, and find command", func(t *testing.T) {
		err := validator(cmd, []string{"17572"})
		requireErrContains(t, err,
			`"17572"`,
			"507f1f77bcf86cd799439011",
			"device id",
			"incloud device list -q 17572",
		)
	})

	t.Run("arity check still enforced", func(t *testing.T) {
		err := validator(cmd, []string{})
		if err == nil {
			t.Fatal("expected an arity error for zero args")
		}
	})

	t.Run("wrong arity reported before id shape", func(t *testing.T) {
		err := validator(cmd, []string{"a", "b"})
		if err == nil {
			t.Fatal("expected an arity error for two args")
		}
		if strings.Contains(err.Error(), "expected a 24-character hex ObjectId") {
			t.Errorf("arity error should not be replaced by id-shape error: %v", err)
		}
	})
}

func TestObjectIDArgsFunc_ScopedFindCommand(t *testing.T) {
	// Simulates a subordinate resource (e.g. project id) looked up within an
	// already-validated parent id (e.g. group id) rather than the offending
	// value itself.
	validator := ObjectIDArgsFunc(cobra.ExactArgs(2), 1, "project id", func(args []string) string {
		return "incloud device group project list " + args[0]
	})
	cmd := &cobra.Command{}

	err := validator(cmd, []string{"507f1f77bcf86cd799439011", "not-an-id"})
	requireErrContains(t, err,
		`"not-an-id"`,
		"project id",
		"incloud device group project list 507f1f77bcf86cd799439011",
	)
}

func TestObjectIDCSVArgs(t *testing.T) {
	validator := ObjectIDCSVArgs(cobra.ExactArgs(1), 0, "device id", "incloud device list -q %s")
	cmd := &cobra.Command{}

	t.Run("all valid ids pass", func(t *testing.T) {
		err := validator(cmd, []string{"507f1f77bcf86cd799439011,653b1ff2a84e171614d88695"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("single valid id passes", func(t *testing.T) {
		if err := validator(cmd, []string{"507f1f77bcf86cd799439011"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("one invalid id among valid ones is reported", func(t *testing.T) {
		err := validator(cmd, []string{"507f1f77bcf86cd799439011,bad-id"})
		requireErrContains(t, err, `"bad-id"`, "incloud device list -q bad-id")
	})
}

func TestObjectIDCSVArgsFunc_ScopedFindCommand(t *testing.T) {
	validator := ObjectIDCSVArgsFunc(cobra.ExactArgs(2), 1, "layerfs id", func(args []string, _ string) string {
		return "incloud device group layerfs list " + args[0]
	})
	cmd := &cobra.Command{}

	err := validator(cmd, []string{"507f1f77bcf86cd799439011", "653b1ff2a84e171614d88695,bad-id"})
	requireErrContains(t, err,
		`"bad-id"`,
		"layerfs id",
		"incloud device group layerfs list 507f1f77bcf86cd799439011",
	)
}

func TestObjectIDArgsAll(t *testing.T) {
	validator := ObjectIDArgsAll(cobra.MinimumNArgs(1), "asset id", "incloud device asset list --name %s")
	cmd := &cobra.Command{}

	t.Run("all valid ids pass", func(t *testing.T) {
		err := validator(cmd, []string{"507f1f77bcf86cd799439011", "653b1ff2a84e171614d88695"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid id among args is reported", func(t *testing.T) {
		err := validator(cmd, []string{"507f1f77bcf86cd799439011", "typo"})
		requireErrContains(t, err, `"typo"`, "asset id", "incloud device asset list --name typo")
	})

	t.Run("arity still enforced", func(t *testing.T) {
		if err := validator(cmd, []string{}); err == nil {
			t.Fatal("expected an arity error for zero args")
		}
	})
}
