package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NoArgsWithHint returns a cobra.PositionalArgs that rejects any positional
// argument, like cobra.NoArgs, but appends a hint built from the first
// offending value. Use it on list commands whose filters are flags, so a
// value passed positionally (e.g. a device id) fails loudly with the right
// flag instead of being silently ignored and returning unfiltered results.
func NoArgsWithHint(hint func(arg string) string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		return fmt.Errorf("unexpected argument %q: %s takes no positional arguments; %s",
			args[0], cmd.CommandPath(), hint(args[0]))
	}
}

// DeviceFilterHint returns a NoArgsWithHint hint for list commands that filter
// by a device id flag: an ObjectId-shaped value is pointed at the flag
// (e.g. "incloud device client list --device <id>"), anything else at
// findCmdTemplate (a template with one %s) to look the id up first.
func DeviceFilterHint(listCmd, flag, findCmdTemplate string) func(arg string) string {
	return func(arg string) string {
		if IsObjectID(arg) {
			return fmt.Sprintf("to filter by device run '%s --%s %s'", listCmd, flag, arg)
		}
		return fmt.Sprintf("to filter by device, run '%s' to find its id, then '%s --%s <id>'",
			FindCmd(findCmdTemplate, arg), listCmd, flag)
	}
}
