package validation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrWebhookVerificationFailed = errors.New("webhook verification failed")

// VerifyMetaWebhookChallenge validates Meta GET webhook subscription handshake.
func VerifyMetaWebhookChallenge(mode, token, challenge, expectedVerifyToken string) (string, error) {
	if mode != "subscribe" || token != expectedVerifyToken {
		return "", ErrWebhookVerificationFailed
	}
	if challenge == "" {
		return "", ErrWebhookVerificationFailed
	}
	return challenge, nil
}

// VerifyMetaSignature validates Meta X-Hub-Signature-256 header against the raw POST body.
// Meta signs the exact request bytes with the Facebook App Secret (not App ID, not verify token).
func VerifyMetaSignature(appSecret, signatureHeader string, body []byte) bool {
	appSecret = strings.TrimSpace(appSecret)
	signatureHeader = strings.TrimSpace(signatureHeader)
	if appSecret == "" || signatureHeader == "" || len(body) == 0 {
		return false
	}

	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}
	expectedHex := strings.TrimPrefix(signatureHeader, prefix)
	expectedMAC, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	computedMAC := mac.Sum(nil)

	return hmac.Equal(computedMAC, expectedMAC)
}

// ComputeMetaSignature returns sha256=... for debugging (e.g. curl tests).
func ComputeMetaSignature(appSecret string, body []byte) string {
	appSecret = strings.TrimSpace(appSecret)
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
