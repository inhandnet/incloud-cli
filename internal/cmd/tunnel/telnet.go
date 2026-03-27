package tunnel

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// Telnet protocol constants
const (
	iac  = 0xff
	dont = 0xfe
	do   = 0xfd
	wont = 0xfc
	will = 0xfb
	sb   = 0xfa
	se   = 0xf0
)

const recvBufSize = 4096

// telnetClient wraps an io.ReadWriter (typically a smux stream) and handles
// telnet protocol negotiation transparently.
type telnetClient struct {
	conn io.ReadWriter
}

// negotiate processes raw telnet data: strips IAC sequences, builds responses.
// Returns cleaned data and response bytes to send back.
func (tc *telnetClient) negotiate(data []byte) (clean []byte, responses []byte) {
	i := 0
	for i < len(data) {
		if data[i] == iac && i+1 < len(data) {
			cmd := data[i+1]
			switch cmd {
			case do, dont, will, wont:
				if i+2 < len(data) {
					opt := data[i+2]
					if cmd == do {
						responses = append(responses, iac, wont, opt)
					} else if cmd == will {
						responses = append(responses, iac, dont, opt)
					}
					i += 3
				} else {
					i = len(data)
				}
			case sb:
				end := -1
				for j := i + 2; j < len(data)-1; j++ {
					if data[j] == iac && data[j+1] == se {
						end = j + 2
						break
					}
				}
				if end == -1 {
					i = len(data)
				} else {
					i = end
				}
			case iac:
				clean = append(clean, iac)
				i += 2
			default:
				i += 2
			}
		} else {
			clean = append(clean, data[i])
			i++
		}
	}
	return
}

// readUntil reads from the connection until one of the regex patterns matches
// the accumulated text, or timeout is reached.
// Returns (accumulated text, matched pattern index). Index is -1 on timeout.
func (tc *telnetClient) readUntil(patterns []*regexp.Regexp, timeout time.Duration) (string, int) {
	deadline := time.Now().Add(timeout)
	var text strings.Builder
	buf := make([]byte, recvBufSize)

	for time.Now().Before(deadline) {
		if dl, ok := tc.conn.(interface{ SetReadDeadline(time.Time) error }); ok {
			remaining := max(time.Until(deadline), 100*time.Millisecond)
			dl.SetReadDeadline(time.Now().Add(remaining))
		}

		n, err := tc.conn.Read(buf)
		if n > 0 {
			clean, responses := tc.negotiate(buf[:n])
			if len(responses) > 0 {
				tc.conn.Write(responses)
			}
			text.Write(clean)

			// Match patterns against ANSI-cleaned text to handle
			// prompts wrapped in escape sequences (e.g. --More-- in reverse video)
			s := text.String()
			cleaned := cleanOutput(s)
			for i, pat := range patterns {
				if pat.FindStringIndex(cleaned) != nil {
					return s, i
				}
			}
		}
		if err != nil {
			if isTimeout(err) {
				continue
			}
			break
		}
	}
	return text.String(), -1
}

// readUntilLiteral reads until one of the literal strings is found.
// Convenience wrapper over readUntil for simple prompt matching.
func (tc *telnetClient) readUntilLiteral(patterns []string, timeout time.Duration) (string, int) {
	regexps := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		regexps[i] = regexp.MustCompile(regexp.QuoteMeta(p))
	}
	return tc.readUntil(regexps, timeout)
}

// write sends text to the connection.
func (tc *telnetClient) write(text string) error {
	_, err := tc.conn.Write([]byte(text))
	return err
}

// isTimeout checks if an error is a timeout.
func isTimeout(err error) bool {
	if ne, ok := err.(interface{ Timeout() bool }); ok {
		return ne.Timeout()
	}
	return false
}

// --- Output cleaning utilities ---

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]|\r|\x08`)

// cleanOutput strips ANSI escape sequences, carriage returns, and backspaces.
func cleanOutput(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

// promptEchoRE matches lines like "8 hostname# cmd" or "14:30:00 hostname# cmd"
var promptEchoRE = regexp.MustCompile(`^(\d[\d:]*\s+)?\S+[#>]\s*`)

// stripCommandEcho removes the command echo line from output.
func stripCommandEcho(output, cmd string) string {
	lines := strings.Split(output, "\n")
	result := make([]string, 0, len(lines))
	cmd = strings.TrimSpace(cmd)
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == cmd {
			continue
		}
		if strings.Contains(stripped, cmd) && promptEchoRE.MatchString(stripped) {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// stripPromptLine removes the trailing prompt line from output.
func stripPromptLine(output string, prompts []*regexp.Regexp) string {
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return output
	}
	lastIdx := len(lines) - 1
	lastLine := lines[lastIdx]
	for _, p := range prompts {
		if p.FindStringIndex(lastLine) != nil {
			result := strings.Join(lines[:lastIdx], "\n")
			if lastIdx > 0 {
				result += "\n"
			}
			return result
		}
	}
	return output
}

// stripMoreArtifacts removes --More-- and the overwrite characters that follow.
var moreRE = regexp.MustCompile(`--More-- *`)

func stripMoreArtifacts(s string) string {
	return moreRE.ReplaceAllString(s, "")
}

// cleanAndStrip applies all output cleaning.
func cleanAndStrip(output, cmd string, prompts []*regexp.Regexp) string {
	output = cleanOutput(output)
	output = stripMoreArtifacts(output)
	output = stripCommandEcho(output, cmd)
	output = stripPromptLine(output, prompts)
	output = strings.TrimRight(output, "\n ")
	if output != "" {
		output += "\n"
	}
	return output
}

// morePattern matches the --More-- pagination prompt.
var morePattern = regexp.MustCompile(`--More--`)

// execINOSCommand sends a command to INOS CLI, handles --More-- pagination,
// and returns the cleaned output.
func execINOSCommand(tc *telnetClient, cmd string, prompts []*regexp.Regexp, timeout time.Duration) (string, error) {
	if err := tc.write(cmd + "\r"); err != nil {
		return "", fmt.Errorf("write command: %w", err)
	}
	time.Sleep(300 * time.Millisecond)

	var full strings.Builder
	allPatterns := append(append([]*regexp.Regexp{}, prompts...), morePattern)

	for {
		text, idx := tc.readUntil(allPatterns, timeout)
		full.WriteString(text)
		if idx == -1 {
			return cleanAndStrip(full.String(), cmd, prompts), fmt.Errorf("timeout waiting for command response")
		}
		if idx == len(prompts) {
			// --More-- : send space to page
			tc.write(" ")
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}

	return cleanAndStrip(full.String(), cmd, prompts), nil
}

// execShellCommand sends a shell command and returns cleaned output.
func execShellCommand(tc *telnetClient, cmd string, prompts []*regexp.Regexp, timeout time.Duration) (string, error) {
	if err := tc.write(cmd + "\r"); err != nil {
		return "", fmt.Errorf("write command: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	text, idx := tc.readUntil(prompts, timeout)
	if idx == -1 {
		return cleanAndStrip(text, cmd, prompts), fmt.Errorf("timeout waiting for command response")
	}
	return cleanAndStrip(text, cmd, prompts), nil
}
