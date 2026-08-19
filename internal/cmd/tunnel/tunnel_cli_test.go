package tunnel

import (
	"errors"
	"io"
	"regexp"
	"testing"
	"time"
)

// scriptedConn replays canned device output chunk by chunk, then reports EOF.
type scriptedConn struct {
	chunks [][]byte
	i      int
}

func (c *scriptedConn) Read(p []byte) (int, error) {
	if c.i >= len(c.chunks) {
		return 0, io.EOF
	}
	n := copy(p, c.chunks[c.i])
	c.i++
	return n, nil
}

func (c *scriptedConn) Write(p []byte) (int, error) { return len(p), nil }

// gbkChinesePrompt is what ER805 running Chinese firmware answers to 'inhand':
// "请输入密码:" encoded in GBK, followed by an ASCII colon.
var gbkChinesePrompt = []byte{0xc7, 0xeb, 0xca, 0xe4, 0xc8, 0xeb, 0xc3, 0xdc, 0xc2, 0xeb, ':'}

func TestAwaitShellEntry(t *testing.T) {
	inosPrompts := []*regexp.Regexp{regexp.MustCompile(`ER805[#>]\s*$`)}
	shellPrompts := []*regexp.Regexp{regexp.MustCompile(DefaultShellPrompt)}

	tests := []struct {
		name    string
		output  []byte
		wantErr bool
	}{
		{
			name:   "busybox prompt means we are in",
			output: []byte("\r\n/ # "),
		},
		{
			name:   "busybox prompt with path",
			output: []byte("\r\n/root # "),
		},
		{
			name:    "english refusal drops back to INOS",
			output:  []byte("\r\nBad password\r\n16:20:01 ER805# "),
			wantErr: true,
		},
		{
			name:    "chinese refusal drops back to INOS",
			output:  append(append([]byte("\r\n"), 0xc3, 0xdc, 0xc2, 0xeb, 0xb4, 0xed, 0xce, 0xf3), []byte("\r\n16:20:01 ER805# ")...),
			wantErr: true,
		},
		{
			name:    "no response at all",
			output:  []byte("\r\n"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &telnetClient{conn: &scriptedConn{chunks: [][]byte{tt.output}}}
			err := awaitShellEntry(tc, inosPrompts, shellPrompts, 500*time.Millisecond)

			if tt.wantErr && err == nil {
				t.Fatalf("awaitShellEntry(%q) = nil, want error", tt.output)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("awaitShellEntry(%q) = %v, want nil", tt.output, err)
			}
		})
	}
}

func TestAwaitShellPasswordPrompt(t *testing.T) {
	tests := []struct {
		name      string
		output    []byte
		wantErr   bool
		wantNoCmd bool
	}{
		{
			name:   "english prompt",
			output: []byte("inhand\r\nInput password:"),
		},
		{
			name:   "english prompt with trailing space",
			output: []byte("inhand\r\nInput password: "),
		},
		{
			name:   "chinese prompt in GBK",
			output: append([]byte("inhand\r\n"), gbkChinesePrompt...),
		},
		{
			name:   "chinese prompt in UTF-8",
			output: []byte("inhand\r\n请输入密码:"),
		},
		{
			name:   "chinese prompt with fullwidth colon",
			output: []byte("inhand\r\n请输入密码："),
		},
		{
			name:      "device without inhand command",
			output:    []byte("inhand\r\n-sh: inhand: not found\r\n"),
			wantErr:   true,
			wantNoCmd: true,
		},
		{
			name:    "command echo only",
			output:  []byte("inhand\r\n"),
			wantErr: true,
		},
		{
			name:    "banner line ending in colon is not a prompt",
			output:  []byte("inhand\r\nWarning:\r\n"),
			wantErr: true,
		},
		{
			name:    "unknown command",
			output:  []byte("inhand\r\n% Unknown command\r\nrouter# "),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &telnetClient{conn: &scriptedConn{chunks: [][]byte{tt.output}}}
			err := awaitShellPasswordPrompt(tc, 500*time.Millisecond)

			if tt.wantErr && err == nil {
				t.Fatalf("awaitShellPasswordPrompt(%q) = nil, want error", tt.output)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("awaitShellPasswordPrompt(%q) = %v, want nil", tt.output, err)
			}
			if got := errors.Is(err, errNoInhandCommand); got != tt.wantNoCmd {
				t.Errorf("errors.Is(err, errNoInhandCommand) = %v, want %v (err = %v)", got, tt.wantNoCmd, err)
			}
		})
	}
}
