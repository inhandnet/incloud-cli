package iostreams

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/muesli/termenv"
)

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

	// Pretty print for TTY
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return string(data)
	}

	if outputMode == "json" {
		// Explicit -o json: pretty but no color
		return buf.String()
	}

	// Default TTY: colorized JSON
	return colorizeJSON(buf.String(), io.TermOutput())
}

var (
	jsonKeyRe    = regexp.MustCompile(`^(\s*)"([^"]+)":`)
	jsonStringRe = regexp.MustCompile(`: "(.*)"(,?)$`)
	jsonNumberRe = regexp.MustCompile(`: (-?\d+\.?\d*)(,?)$`)
	jsonBoolRe   = regexp.MustCompile(`: (true|false|null)(,?)$`)
)

func colorizeJSON(s string, out *termenv.Output) string {
	c := NewColorizer(out)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		// Colorize keys
		line = jsonKeyRe.ReplaceAllStringFunc(line, func(m string) string {
			parts := jsonKeyRe.FindStringSubmatch(m)
			if len(parts) >= 3 {
				return parts[1] + c.Bold("\""+parts[2]+"\"") + ":"
			}
			return m
		})
		// Colorize string values
		line = jsonStringRe.ReplaceAllStringFunc(line, func(m string) string {
			parts := jsonStringRe.FindStringSubmatch(m)
			if len(parts) >= 3 {
				return ": " + c.Green("\""+parts[1]+"\"") + parts[2]
			}
			return m
		})
		// Colorize number values
		line = jsonNumberRe.ReplaceAllStringFunc(line, func(m string) string {
			parts := jsonNumberRe.FindStringSubmatch(m)
			if len(parts) >= 3 {
				return ": " + c.Yellow(parts[1]) + parts[2]
			}
			return m
		})
		// Colorize bool/null values
		line = jsonBoolRe.ReplaceAllStringFunc(line, func(m string) string {
			parts := jsonBoolRe.FindStringSubmatch(m)
			if len(parts) >= 3 {
				return ": " + c.Red(parts[1]) + parts[2]
			}
			return m
		})
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
