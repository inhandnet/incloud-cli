package iostreams

import (
	"bytes"
	"encoding/json"

	"github.com/tidwall/pretty"
)

// jsonStyle matches the original color scheme: bold keys, green strings,
// yellow numbers, red booleans/null.
var jsonStyle = &pretty.Style{
	Key:    [2]string{"\x1b[1m", "\x1b[0m"},  // bold
	String: [2]string{"\x1b[32m", "\x1b[0m"}, // green
	Number: [2]string{"\x1b[33m", "\x1b[0m"}, // yellow
	True:   [2]string{"\x1b[31m", "\x1b[0m"}, // red
	False:  [2]string{"\x1b[31m", "\x1b[0m"}, // red
	Null:   [2]string{"\x1b[31m", "\x1b[0m"}, // red
	Escape: [2]string{"\x1b[35m", "\x1b[0m"}, // magenta
}

// FormatJSON formats JSON bytes based on TTY state and output mode.
// - TTY default: colorized pretty-print
// - TTY + "json": plain pretty-print (for explicit -o json)
// - Non-TTY: compact JSON (single line)
// - Non-TTY + "json": compact JSON
func FormatJSON(data []byte, io *IOStreams, outputMode string) string {
	if !io.IsStdoutTTY() {
		// Compact JSON for piping
		var buf bytes.Buffer
		if err := json.Compact(&buf, data); err != nil {
			return string(data)
		}
		return buf.String()
	}

	if outputMode == "json" {
		// Explicit -o json: pretty but no color
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return string(data)
		}
		return buf.String()
	}

	// Default TTY: pretty-print + colorize
	if !json.Valid(data) {
		return string(data)
	}
	result := pretty.Pretty(data)
	return string(pretty.Color(result, jsonStyle))
}
