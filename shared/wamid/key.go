package wamid

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

const minMessageKeyLen = 16
const maxStoredKeyLen = 32

// ExtractMessageKey returns the stable WhatsApp message token embedded in a wamid.
// Meta may send different wamid prefixes for the same message (phone vs user_id),
// but the trailing hex token inside the decoded payload stays the same.
//
// Non-WhatsApp ids (e.g. Instagram mid) are stored as a lowercased value when short,
// or a 32-char SHA-256 prefix when longer than VARCHAR(32) wamid_key.
func ExtractMessageKey(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "wamid.") {
		return shortenExternalKey(strings.ToLower(raw))
	}

	body := strings.TrimPrefix(raw, "wamid.")
	decoded, err := base64.StdEncoding.DecodeString(padBase64(body))
	if err != nil {
		return ""
	}
	return extractHexToken(decoded)
}

func shortenExternalKey(raw string) string {
	if len(raw) <= maxStoredKeyLen {
		return raw
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

// KeysMatch reports whether two wamid values refer to the same WhatsApp message.
func KeysMatch(a, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a == "" || b == "" {
		return false
	}
	if strings.EqualFold(a, b) {
		return true
	}
	keyA := ExtractMessageKey(a)
	keyB := ExtractMessageKey(b)
	return keyA != "" && keyA == keyB
}

func padBase64(input string) string {
	mod := len(input) % 4
	if mod == 0 {
		return input
	}
	return input + strings.Repeat("=", 4-mod)
}

func extractHexToken(decoded []byte) string {
	end := len(decoded)
	for end > 0 && decoded[end-1] == 0 {
		end--
	}
	start := end
	for start > 0 && isHexByte(decoded[start-1]) {
		start--
	}
	if end-start < minMessageKeyLen {
		return ""
	}
	return strings.ToUpper(string(decoded[start:end]))
}

func isHexByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'F') || (b >= 'a' && b <= 'f')
}
