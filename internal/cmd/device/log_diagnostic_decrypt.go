package device

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	diagDecryptKey      = "nvram_key@inhand"
	diagBlockPlainMax   = 512
	diagAESHeaderPrefix = "$AES$"
)

// decryptDiagnosticLog decrypts an AES-128-CBC encrypted diagnostic log.
//
// File format:
//
//	Line 1:  $AES$ len=<plaintext-length>
//	Line 2+: hex-encoded ciphertext (whitespace ignored)
//
// The device encrypts in 512-byte plaintext chunks, each independently
// AES-128-CBC encrypted with IV=0 and no padding. The last chunk may be
// shorter than 512 bytes but is padded to a 16-byte boundary for encryption.
func decryptDiagnosticLog(data []byte) ([]byte, error) {
	text := string(data)

	// Split header from body.
	headerLine, body, found := strings.Cut(text, "\n")
	if !found {
		return nil, fmt.Errorf("invalid diagnostic log: missing header line")
	}
	header := strings.TrimSpace(headerLine)

	// Parse header: $AES$ len=N
	if !strings.HasPrefix(header, diagAESHeaderPrefix) {
		return nil, fmt.Errorf("invalid diagnostic log: header %q does not start with %s", header, diagAESHeaderPrefix)
	}
	lenPart := strings.TrimPrefix(header, diagAESHeaderPrefix)
	lenPart = strings.TrimSpace(lenPart)
	if !strings.HasPrefix(lenPart, "len=") {
		return nil, fmt.Errorf("invalid diagnostic log: header %q missing len= field", header)
	}
	plainLen, err := strconv.Atoi(strings.TrimPrefix(lenPart, "len="))
	if err != nil {
		return nil, fmt.Errorf("invalid diagnostic log: bad length in header: %w", err)
	}

	// Strip all whitespace from hex body and decode.
	hexStr := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, body)
	cipherData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("decoding hex ciphertext: %w", err)
	}

	// Prepare AES cipher with zero IV.
	key := make([]byte, aes.BlockSize)
	copy(key, diagDecryptKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}
	iv := make([]byte, aes.BlockSize)

	// Decrypt in 512-byte plaintext chunks.
	out := make([]byte, 0, plainLen)
	remaining := plainLen
	offset := 0

	for remaining > 0 {
		chunkPlain := min(remaining, diagBlockPlainMax)
		// Ciphertext is padded to 16-byte boundary.
		chunkCipher := ((chunkPlain + aes.BlockSize - 1) / aes.BlockSize) * aes.BlockSize

		if offset+chunkCipher > len(cipherData) {
			return nil, fmt.Errorf("ciphertext too short: need %d bytes at offset %d, have %d total",
				chunkCipher, offset, len(cipherData))
		}

		chunk := make([]byte, chunkCipher)
		copy(chunk, cipherData[offset:offset+chunkCipher])

		// Each chunk uses a fresh zero IV.
		decrypter := cipher.NewCBCDecrypter(block, iv)
		decrypter.CryptBlocks(chunk, chunk)

		out = append(out, chunk[:chunkPlain]...)
		offset += chunkCipher
		remaining -= chunkPlain
	}

	// Sanity check: decrypted output should be a gzip file (magic: 0x1f 0x8b).
	if len(out) >= 2 && (out[0] != 0x1f || out[1] != 0x8b) {
		return nil, fmt.Errorf("decrypted data is not a valid gzip file (wrong magic bytes: %02x %02x); possibly wrong decryption key", out[0], out[1])
	}

	return out, nil
}

// isDiagnosticEncrypted checks whether the data starts with the $AES$ header.
func isDiagnosticEncrypted(data []byte) bool {
	return len(data) > len(diagAESHeaderPrefix) &&
		string(data[:len(diagAESHeaderPrefix)]) == diagAESHeaderPrefix
}
