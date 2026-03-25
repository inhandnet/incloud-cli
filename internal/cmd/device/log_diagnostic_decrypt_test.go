package device

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

// gzipMagic is prepended to test plaintext so decrypted output passes the gzip check.
var gzipMagic = []byte{0x1f, 0x8b}

// encryptDiagnosticLog is a test helper that produces the $AES$ format.
func encryptDiagnosticLog(plain []byte) []byte {
	key := make([]byte, aes.BlockSize)
	copy(key, diagDecryptKey)
	block, _ := aes.NewCipher(key)
	iv := make([]byte, aes.BlockSize)

	var cipherHex []byte
	remaining := len(plain)
	offset := 0

	for remaining > 0 {
		chunkPlain := min(remaining, diagBlockPlainMax)
		chunkCipher := ((chunkPlain + aes.BlockSize - 1) / aes.BlockSize) * aes.BlockSize

		buf := make([]byte, chunkCipher)
		copy(buf, plain[offset:offset+chunkPlain])

		enc := cipher.NewCBCEncrypter(block, iv)
		enc.CryptBlocks(buf, buf)

		cipherHex = append(cipherHex, []byte(hex.EncodeToString(buf))...)
		offset += chunkPlain
		remaining -= chunkPlain
	}

	header := fmt.Sprintf("$AES$ len=%d\n", len(plain))
	return append([]byte(header), cipherHex...)
}

// makeGzipPayload creates test plaintext with gzip magic header prefix.
func makeGzipPayload(size int) []byte {
	buf := make([]byte, size)
	buf[0] = 0x1f
	buf[1] = 0x8b
	for i := 2; i < size; i++ {
		buf[i] = byte(i % 251)
	}
	return buf
}

func TestDecryptDiagnosticLog(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"small (< 16 bytes)", append(gzipMagic, []byte("hello wor")...)},
		{"exact block (16 bytes)", makeGzipPayload(16)},
		{"one chunk (512 bytes)", makeGzipPayload(512)},
		{"multi chunk (1500 bytes)", makeGzipPayload(1500)},
		{"non-aligned (100 bytes)", makeGzipPayload(100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encryptDiagnosticLog(tt.plaintext)
			decrypted, err := decryptDiagnosticLog(encrypted)
			if err != nil {
				t.Fatalf("decryptDiagnosticLog() error = %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Fatalf("length %d vs %d, content mismatch", len(decrypted), len(tt.plaintext))
			}
		})
	}
}

func TestDecryptDiagnosticLogWithWhitespace(t *testing.T) {
	plain := makeGzipPayload(16)
	encrypted := encryptDiagnosticLog(plain)

	// Insert whitespace into the hex body.
	text := string(encrypted)
	parts := []byte(text)
	for i := len("$AES$ len=16\n") + 8; i < len(parts); i += 9 {
		parts = append(parts[:i], append([]byte(" \n "), parts[i:]...)...)
	}

	decrypted, err := decryptDiagnosticLog(parts)
	if err != nil {
		t.Fatalf("decryptDiagnosticLog() error = %v", err)
	}

	if !bytes.Equal(decrypted, plain) {
		t.Fatalf("got %q, want %q", decrypted, plain)
	}
}

func TestDecryptDiagnosticLogErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr string
	}{
		{
			name:    "no newline",
			input:   []byte("$AES$ len=10"),
			wantErr: "missing header line",
		},
		{
			name:    "bad header prefix",
			input:   []byte("NOTAES len=10\nhex"),
			wantErr: "does not start with",
		},
		{
			name:    "missing len field",
			input:   []byte("$AES$ size=10\nhex"),
			wantErr: "missing len= field",
		},
		{
			name:    "non-numeric length",
			input:   []byte("$AES$ len=abc\nhex"),
			wantErr: "bad length in header",
		},
		{
			name:    "invalid hex",
			input:   []byte("$AES$ len=16\nZZZZ"),
			wantErr: "decoding hex ciphertext",
		},
		{
			name:    "ciphertext too short",
			input:   []byte("$AES$ len=512\nabcd"),
			wantErr: "ciphertext too short",
		},
		{
			name:    "bad gzip magic (wrong key scenario)",
			input:   encryptDiagnosticLog([]byte("not a gzip file!")),
			wantErr: "not a valid gzip file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decryptDiagnosticLog(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestIsDiagnosticEncrypted(t *testing.T) {
	if !isDiagnosticEncrypted([]byte("$AES$ len=100\nabcdef")) {
		t.Error("expected true for $AES$ header")
	}
	if isDiagnosticEncrypted([]byte("plain text log")) {
		t.Error("expected false for plain text")
	}
	if isDiagnosticEncrypted([]byte("$AES")) {
		t.Error("expected false for truncated header")
	}
}
