package debug

import (
	"fmt"
	"os"
)

// Enabled controls whether debug output is printed.
// Set by --debug flag or INCLOUD_DEBUG env var in root command.
var Enabled bool

// Log prints a debug message to stderr with [debug] prefix.
func Log(format string, args ...any) {
	if Enabled {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}
