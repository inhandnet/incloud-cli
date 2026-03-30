package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
)

// newTestCmd creates a cobra command with the given flags for testing.
func newTestCmd(flags func(cmd *cobra.Command)) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	flags(cmd)
	return cmd
}

func TestNewQuery_AllFlagsSet(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().String("sort", "", "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().StringSlice("expand", nil, "")
	})
	// Simulate user passing all flags
	cmd.SetArgs([]string{
		"--page", "3",
		"--limit", "50",
		"--sort", "name,asc",
		"--fields", "a,b",
		"--expand", "org,firmware",
	})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)

	if got := q.Get("page"); got != "2" {
		t.Errorf("page: got %q, want %q", got, "2")
	}
	if got := q.Get("limit"); got != "50" {
		t.Errorf("limit: got %q, want %q", got, "50")
	}
	if got := q.Get("sort"); got != "name,asc" {
		t.Errorf("sort: got %q, want %q", got, "name,asc")
	}
	if got := q.Get("fields"); got != "a,b" {
		t.Errorf("fields: got %q, want %q", got, "a,b")
	}
	if got := q.Get("expand"); got != "org,firmware" {
		t.Errorf("expand: got %q, want %q", got, "org,firmware")
	}
}

func TestNewQuery_PageConversion(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
	})
	cmd.SetArgs([]string{"--page", "1"})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)
	if got := q.Get("page"); got != "0" {
		t.Errorf("page 1 should convert to 0, got %q", got)
	}
}

func TestNewQuery_DefaultValues(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().String("sort", "", "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().StringSlice("expand", nil, "")
	})
	// No args — use all defaults
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)

	// page and limit should still be set (with defaults)
	if got := q.Get("page"); got != "0" {
		t.Errorf("page: got %q, want %q", got, "0")
	}
	if got := q.Get("limit"); got != "20" {
		t.Errorf("limit: got %q, want %q", got, "20")
	}

	// sort, fields should NOT be set (not Changed)
	if got := q.Get("sort"); got != "" {
		t.Errorf("sort should be empty when not changed, got %q", got)
	}
	if got := q.Get("fields"); got != "" {
		t.Errorf("fields should be empty when not changed, got %q", got)
	}

	// expand should NOT be set (nil default, not non-empty)
	if got := q.Get("expand"); got != "" {
		t.Errorf("expand should be empty when nil default, got %q", got)
	}
}

func TestNewQuery_MissingFlags(t *testing.T) {
	// Command with only page and limit — no sort/fields/expand
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 10, "")
	})
	cmd.SetArgs([]string{"--page", "2"})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)

	if got := q.Get("page"); got != "1" {
		t.Errorf("page: got %q, want %q", got, "1")
	}
	if got := q.Get("limit"); got != "10" {
		t.Errorf("limit: got %q, want %q", got, "10")
	}
	// Missing flags should not appear
	if _, ok := q["sort"]; ok {
		t.Error("sort should not be present")
	}
	if _, ok := q["fields"]; ok {
		t.Error("fields should not be present")
	}
	if _, ok := q["expand"]; ok {
		t.Error("expand should not be present")
	}
}

func TestNewQuery_FieldsWildcard(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().StringSlice("fields", nil, "")
	})
	cmd.SetArgs([]string{"--fields", "*"})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)
	if got := q.Get("fields"); got != "*" {
		t.Errorf("fields wildcard: got %q, want %q", got, "*")
	}
}

func TestNewQuery_ExpandWithDefault(t *testing.T) {
	// expand has a non-empty default but user didn't pass the flag
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().StringSlice("expand", []string{"org"}, "")
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)
	// Should be set because the value is non-empty, even though not Changed
	if got := q.Get("expand"); got != "org" {
		t.Errorf("expand with default: got %q, want %q", got, "org")
	}
}

func TestNewQuery_ExpandOverrideDefault(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().StringSlice("expand", []string{"org"}, "")
	})
	cmd.SetArgs([]string{"--expand", "firmware,status"})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)
	if got := q.Get("expand"); got != "firmware,status" {
		t.Errorf("expand override: got %q, want %q", got, "firmware,status")
	}
}

func TestNewQuery_DefaultFieldsApplied(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	defaults := []string{"_id", "name", "online"}
	q := NewQuery(cmd, defaults)

	// output defaults to "" which is treated as table → defaultFields applied
	if got := q.Get("fields"); got != "_id,name,online" {
		t.Errorf("default fields: got %q, want %q", got, "_id,name,online")
	}
}

func TestNewQuery_DefaultFieldsNotAppliedForJSON(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{"--output", "json"})
	_ = cmd.Execute()

	defaults := []string{"_id", "name", "online"}
	q := NewQuery(cmd, defaults)

	// json output → defaultFields NOT applied
	if got := q.Get("fields"); got != "" {
		t.Errorf("default fields should not apply for json, got %q", got)
	}
}

func TestNewQuery_UserFieldsOverrideDefaults(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{"--fields", "x,y"})
	_ = cmd.Execute()

	defaults := []string{"_id", "name", "online"}
	q := NewQuery(cmd, defaults)

	// user-specified fields override defaults
	if got := q.Get("fields"); got != "x,y" {
		t.Errorf("user fields: got %q, want %q", got, "x,y")
	}
}

func TestNewQuery_ExplicitTableOutput(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{"--output", "table"})
	_ = cmd.Execute()

	defaults := []string{"_id", "name"}
	q := NewQuery(cmd, defaults)

	// explicit --output table should apply default fields
	if got := q.Get("fields"); got != "_id,name" {
		t.Errorf("fields with explicit table: got %q, want %q", got, "_id,name")
	}
}

func TestNewQuery_YAMLOutputNoDefaultFields(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{"--output", "yaml"})
	_ = cmd.Execute()

	defaults := []string{"_id", "name"}
	q := NewQuery(cmd, defaults)

	// yaml output should NOT apply default fields
	if got := q.Get("fields"); got != "" {
		t.Errorf("fields with yaml output should be empty, got %q", got)
	}
}

func TestNewQuery_NilDefaultFieldsNoFieldsSet(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)

	// nil defaults + no user fields = no fields param
	if _, ok := q["fields"]; ok {
		t.Error("fields should not be present when defaults are nil and user didn't specify")
	}
}

func TestNewQuery_EmptyDefaultFieldsNoFieldsSet(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		cmd.Flags().String("output", "", "")
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	q := NewQuery(cmd, []string{})

	// empty defaults + no user fields = no fields param
	if _, ok := q["fields"]; ok {
		t.Error("fields should not be present when defaults are empty and user didn't specify")
	}
}

func TestNewQuery_SortNotSetWhenUnchanged(t *testing.T) {
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().String("sort", "createdAt,desc", "") // non-empty default
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	q := NewQuery(cmd, nil)

	// sort has a non-empty default but was not Changed — should NOT be set
	if got := q.Get("sort"); got != "" {
		t.Errorf("sort should be empty when not changed, got %q", got)
	}
}

func TestNewQuery_NoOutputFlag(t *testing.T) {
	// Command without --output flag (like some get commands)
	cmd := newTestCmd(func(cmd *cobra.Command) {
		cmd.Flags().Int("page", 1, "")
		cmd.Flags().Int("limit", 20, "")
		cmd.Flags().StringSlice("fields", nil, "")
		// No --output flag
	})
	cmd.SetArgs([]string{})
	_ = cmd.Execute()

	defaults := []string{"_id", "name"}
	q := NewQuery(cmd, defaults)

	// When --output flag is missing, GetString returns error, so output is ""
	// which is treated as table → default fields should apply
	if got := q.Get("fields"); got != "_id,name" {
		t.Errorf("fields without output flag: got %q, want %q", got, "_id,name")
	}
}

func TestListFlags_Register(t *testing.T) {
	lf := &ListFlags{}
	cmd := &cobra.Command{Use: "test"}
	lf.Register(cmd)

	cmd.SetArgs([]string{"--page", "2", "--limit", "50", "--sort", "name,asc", "--fields", "a,b"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if lf.Page != 2 {
		t.Errorf("Page: got %d, want 2", lf.Page)
	}
	if lf.Limit != 50 {
		t.Errorf("Limit: got %d, want 50", lf.Limit)
	}
	if lf.Sort != "name,asc" {
		t.Errorf("Sort: got %q, want %q", lf.Sort, "name,asc")
	}
	if len(lf.Fields) != 2 || lf.Fields[0] != "a" || lf.Fields[1] != "b" {
		t.Errorf("Fields: got %v, want [a b]", lf.Fields)
	}
}

func TestListFlags_Register_Defaults(t *testing.T) {
	lf := &ListFlags{}
	cmd := &cobra.Command{Use: "test"}
	lf.Register(cmd)

	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if lf.Page != 1 {
		t.Errorf("Page default: got %d, want 1", lf.Page)
	}
	if lf.Limit != 20 {
		t.Errorf("Limit default: got %d, want 20", lf.Limit)
	}
}

func TestListFlags_RegisterExpand(t *testing.T) {
	lf := &ListFlags{}
	cmd := &cobra.Command{Use: "test"}
	lf.RegisterExpand(cmd)

	cmd.SetArgs([]string{"--expand", "org,firmware"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if len(lf.Expand) != 2 || lf.Expand[0] != "org" || lf.Expand[1] != "firmware" {
		t.Errorf("Expand: got %v, want [org firmware]", lf.Expand)
	}
}
