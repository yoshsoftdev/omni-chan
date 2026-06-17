package normalizer

import (
	"testing"

	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

const capturedMetaWebhookBody = `{"object":"whatsapp_business_account","entry":[{"id":"102272935850838","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"6282113389840","phone_number_id":"106996032038671"},"contacts":[{"profile":{"name":"YoShandy"},"wa_id":"6282288812003","user_id":"ID.37030824809838146"}],"messages":[{"from":"6282288812003","from_user_id":"ID.37030824809838146","id":"wamid.HBgNNjI4MjI4ODgxMjAwMxUCABIYIEFDQzk2QUQ0RjNERTg3RDAwRURGQzE1NTlDNDUwRTNGAA==","timestamp":"1781251089","text":{"body":"Halo"},"type":"text"}]},"field":"messages"}]}]}`

func TestNormalizeMeta_CapturedInboundMessage(t *testing.T) {
	events, err := NormalizeMeta([]byte(capturedMetaWebhookBody))
	if err != nil {
		t.Fatalf("NormalizeMeta() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != sharedevents.WhatsAppMessageReceived {
		t.Fatalf("type = %s", events[0].Type)
	}

	payload, ok := events[0].Payload.(sharedevents.WhatsAppMessageReceivedPayload)
	if !ok {
		t.Fatal("unexpected payload type")
	}
	if payload.PhoneNumberID != "106996032038671" {
		t.Fatalf("phone_number_id = %q", payload.PhoneNumberID)
	}
	if payload.From != "6282288812003" {
		t.Fatalf("from = %q", payload.From)
	}
	if payload.Content != "Halo" {
		t.Fatalf("content = %q", payload.Content)
	}
	if payload.ContactName != "YoShandy" {
		t.Fatalf("contact_name = %q", payload.ContactName)
	}
	if payload.WABAID != "102272935850838" {
		t.Fatalf("waba_id = %q", payload.WABAID)
	}
	if payload.WAMID == "" {
		t.Fatal("expected wamid")
	}
}
