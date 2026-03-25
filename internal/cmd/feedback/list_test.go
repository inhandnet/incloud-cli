package feedback

import "testing"

func TestFormatAttachments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"empty brackets", "[]", ""},
		{"brackets with space", "[ ]", ""},
		{"single empty string in array", "[]", ""},
		{"single attachment", "[2026-03-25/abc123/file.png]", "file.png"},
		{"multiple attachments", "[2026-03-25/abc123/file.png 2026-03-26/def456/doc.pdf]", "file.png, doc.pdf"},
		{"attachment with deep path", "[a/b/c/d/file.txt]", "file.txt"},
		{"filename only", "[readme.md]", "readme.md"},
		{"mixed empty and valid", "[2026-03-25/abc/file.png]", "file.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAttachments(tt.input)
			if got != tt.want {
				t.Errorf("formatAttachments(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
