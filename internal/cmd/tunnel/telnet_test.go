package tunnel

import (
	"regexp"
	"testing"
)

func TestNegotiateDO(t *testing.T) {
	tc := &telnetClient{}
	input := []byte{0xff, 0xfd, 0x01, 'h', 'i'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "hi" {
		t.Errorf("clean = %q, want %q", clean, "hi")
	}
	if len(responses) != 3 || responses[0] != 0xff || responses[1] != 0xfc || responses[2] != 0x01 {
		t.Errorf("responses = %x, want ff fc 01", responses)
	}
}

func TestNegotiateWILL(t *testing.T) {
	tc := &telnetClient{}
	input := []byte{0xff, 0xfb, 0x03, 'o', 'k'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "ok" {
		t.Errorf("clean = %q, want %q", clean, "ok")
	}
	if len(responses) != 3 || responses[0] != 0xff || responses[1] != 0xfe || responses[2] != 0x03 {
		t.Errorf("responses = %x, want ff fe 03", responses)
	}
}

func TestNegotiateDONT(t *testing.T) {
	tc := &telnetClient{}
	// DONT should be stripped with no response
	input := []byte{0xff, 0xfe, 0x01, 'x'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "x" {
		t.Errorf("clean = %q, want %q", clean, "x")
	}
	if len(responses) != 0 {
		t.Errorf("responses should be empty for DONT, got %x", responses)
	}
}

func TestNegotiateWONT(t *testing.T) {
	tc := &telnetClient{}
	input := []byte{0xff, 0xfc, 0x01, 'y'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "y" {
		t.Errorf("clean = %q, want %q", clean, "y")
	}
	if len(responses) != 0 {
		t.Errorf("responses should be empty for WONT, got %x", responses)
	}
}

func TestNegotiateSubnegotiation(t *testing.T) {
	tc := &telnetClient{}
	input := []byte{0xff, 0xfa, 0x01, 0x02, 0xff, 0xf0, 'x'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "x" {
		t.Errorf("clean = %q, want %q", clean, "x")
	}
	if len(responses) != 0 {
		t.Errorf("responses should be empty for SB, got %x", responses)
	}
}

func TestNegotiateEscapedIAC(t *testing.T) {
	tc := &telnetClient{}
	input := []byte{0xff, 0xff, 'a'}
	clean, responses := tc.negotiate(input)
	if len(clean) != 2 || clean[0] != 0xff || clean[1] != 'a' {
		t.Errorf("clean = %x, want ff 61", clean)
	}
	if len(responses) != 0 {
		t.Errorf("responses should be empty")
	}
}

func TestNegotiateMultipleCommands(t *testing.T) {
	tc := &telnetClient{}
	// IAC DO 0x01 + IAC WILL 0x03 + "hi"
	input := []byte{0xff, 0xfd, 0x01, 0xff, 0xfb, 0x03, 'h', 'i'}
	clean, responses := tc.negotiate(input)
	if string(clean) != "hi" {
		t.Errorf("clean = %q, want %q", clean, "hi")
	}
	// Should have WONT 0x01 + DONT 0x03
	expected := []byte{0xff, 0xfc, 0x01, 0xff, 0xfe, 0x03}
	if len(responses) != len(expected) {
		t.Errorf("responses = %x, want %x", responses, expected)
	} else {
		for i := range expected {
			if responses[i] != expected[i] {
				t.Errorf("responses[%d] = %x, want %x", i, responses[i], expected[i])
			}
		}
	}
}

func TestNegotiatePureText(t *testing.T) {
	tc := &telnetClient{}
	input := []byte("hello world")
	clean, responses := tc.negotiate(input)
	if string(clean) != "hello world" {
		t.Errorf("clean = %q, want %q", clean, "hello world")
	}
	if len(responses) != 0 {
		t.Errorf("responses should be empty for pure text")
	}
}

func TestCleanOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ANSI color", "\x1b[32mhello\x1b[0m", "hello"},
		{"carriage return", "line1\r\nline2\r\n", "line1\nline2\n"},
		{"mixed", "\x1b[1;33mwarn\x1b[0m\r\n", "warn\n"},
		{"plain text", "plain text", "plain text"},
		{"cursor movement", "\x1b[24;27Htext", "text"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanOutput(tt.input)
			if got != tt.want {
				t.Errorf("cleanOutput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripCommandEcho(t *testing.T) {
	tests := []struct {
		name   string
		output string
		cmd    string
		want   string
	}{
		{
			name:   "plain echo",
			output: "show interface\neth0: UP\n",
			cmd:    "show interface",
			want:   "eth0: UP\n",
		},
		{
			name:   "echo with prompt prefix",
			output: "8 er# show interface\neth0: UP\n",
			cmd:    "show interface",
			want:   "eth0: UP\n",
		},
		{
			name:   "timestamp prompt prefix",
			output: "14:30:00 router# show interface\neth0: UP\n",
			cmd:    "show interface",
			want:   "eth0: UP\n",
		},
		{
			name:   "no echo",
			output: "eth0: UP\n",
			cmd:    "show interface",
			want:   "eth0: UP\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCommandEcho(tt.output, tt.cmd)
			if got != tt.want {
				t.Errorf("stripCommandEcho() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripPromptLine(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		prompts []*regexp.Regexp
		want    string
	}{
		{
			name:    "hostname prompt",
			output:  "eth0: UP\n14:30:00 router# ",
			prompts: []*regexp.Regexp{regexp.MustCompile(`router[#>]`)},
			want:    "eth0: UP\n",
		},
		{
			name:    "numbered prompt",
			output:  "eth0: UP\n5 router# ",
			prompts: []*regexp.Regexp{regexp.MustCompile(`router[#>]`)},
			want:    "eth0: UP\n",
		},
		{
			name:    "no match",
			output:  "result\n",
			prompts: []*regexp.Regexp{regexp.MustCompile(`router[#>]`)},
			want:    "result\n",
		},
		{
			name:    "shell prompt",
			output:  "total 4\ndrwxr-xr-x 2 root root 40 Jan  1 00:00 tmp\n/www # ",
			prompts: []*regexp.Regexp{regexp.MustCompile(`\S+\s*[#$]\s*$`)},
			want:    "total 4\ndrwxr-xr-x 2 root root 40 Jan  1 00:00 tmp\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripPromptLine(tt.output, tt.prompts)
			if got != tt.want {
				t.Errorf("stripPromptLine(%q) = %q, want %q", tt.output, got, tt.want)
			}
		})
	}
}

func TestStripMoreArtifacts(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"line1\n--More-- \nline2", "line1\n\nline2"},
		{"--More--  data", "data"},
		{"no paging here", "no paging here"},
	}
	for _, tt := range tests {
		got := stripMoreArtifacts(tt.input)
		if got != tt.want {
			t.Errorf("stripMoreArtifacts(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCleanAndStrip(t *testing.T) {
	prompts := []*regexp.Regexp{regexp.MustCompile(`router[#>]`)}
	output := "show log\r\n\x1b[32mJan 1 00:00:00 syslog: test\x1b[0m\r\n14:30:00 router# "
	got := cleanAndStrip(output, "show log", prompts)
	want := "Jan 1 00:00:00 syslog: test\n"
	if got != want {
		t.Errorf("cleanAndStrip() = %q, want %q", got, want)
	}
}
