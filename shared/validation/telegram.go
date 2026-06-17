package validation

import (
	"crypto/subtle"
	"strings"
)

// VerifyTelegramSecretToken validates X-Telegram-Bot-Api-Secret-Token header.
func VerifyTelegramSecretToken(expected, provided string) bool {
	expected = strings.TrimSpace(expected)
	provided = strings.TrimSpace(provided)
	if expected == "" || provided == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(provided)) == 1
}
