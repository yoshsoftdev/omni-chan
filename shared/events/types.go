package events

import (
	"encoding/json"
	"time"
)

// Event type constants for webhook normalization.
const (
	WhatsAppMessageReceived  = "whatsapp.message.received"
	WhatsAppMessageStatus    = "whatsapp.message.status"
	WhatsAppTemplateStatus   = "whatsapp.template.status"
	WhatsAppAccountUpdate    = "whatsapp.account.update"
	TelegramMessageReceived  = "telegram.message.received"
	TelegramCallbackReceived = "telegram.callback.received"
)

const (
	SourceMeta     = "meta"
	SourceTelegram = "telegram"
)

// Envelope is the canonical internal event format published to NATS.
type Envelope struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Timestamp time.Time         `json:"timestamp"`
	Payload   json.RawMessage   `json:"payload"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Subject returns the NATS subject for this event type.
func Subject(eventType string) string {
	return "events." + eventType
}
