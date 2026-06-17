package validation

import "testing"

// Captured from Meta webhook study log (2026-06-12).
const capturedMetaWebhookBody = `{"object":"whatsapp_business_account","entry":[{"id":"102272935850838","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"6282113389840","phone_number_id":"106996032038671"},"contacts":[{"profile":{"name":"YoShandy"},"wa_id":"6282288812003","user_id":"ID.37030824809838146"}],"messages":[{"from":"6282288812003","from_user_id":"ID.37030824809838146","id":"wamid.HBgNNjI4MjI4ODgxMjAwMxUCABIYIEFDQzk2QUQ0RjNERTg3RDAwRURGQzE1NTlDNDUwRTNGAA==","timestamp":"1781251089","text":{"body":"Halo"},"type":"text"}]},"field":"messages"}]}]}`

const capturedMetaSignature = "sha256=bf37a954cbf6b10ab53ebbfa322e88989d606dc240225d992e250722b1766923"

func TestVerifyMetaSignature_RoundTrip(t *testing.T) {
	body := []byte(capturedMetaWebhookBody)
	secret := "test-app-secret"

	sig := ComputeMetaSignature(secret, body)
	if !VerifyMetaSignature(secret, sig, body) {
		t.Fatal("expected valid signature for computed header")
	}
}

func TestVerifyMetaSignature_CapturedMetaPayloadRejectsWrongSecret(t *testing.T) {
	body := []byte(capturedMetaWebhookBody)
	wrongSecret := "d9fa810caf1c8f282d2d219625ec2694"

	if VerifyMetaSignature(wrongSecret, capturedMetaSignature, body) {
		t.Fatal("expected verification to fail when app secret does not match Meta's signing key")
	}

	computed := ComputeMetaSignature(wrongSecret, body)
	if computed == capturedMetaSignature {
		t.Fatal("test fixture is inconsistent")
	}
}

func TestVerifyMetaSignature_RejectsTamperedBody(t *testing.T) {
	body := []byte(capturedMetaWebhookBody)
	secret := "test-app-secret"
	sig := ComputeMetaSignature(secret, body)

	tampered := append(append([]byte{}, body...), ' ')
	if VerifyMetaSignature(secret, sig, tampered) {
		t.Fatal("expected tampered body to fail verification")
	}
}
