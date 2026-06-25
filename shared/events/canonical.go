package events

import (
	"encoding/json"
	"time"
)

// Provider identifiers for canonical events.
const (
	ProviderMeta     = "meta"
	ProviderTelegram = "telegram"
)

// Canonical event type strings (provider-neutral).
const (
	TypeMessageReceived       = "message.received"
	TypeMessageStatusSent     = "message.status.sent"
	TypeMessageStatusDelivered = "message.status.delivered"
	TypeMessageStatusRead     = "message.status.read"
	TypeMessageStatusFailed   = "message.status.failed"
	TypeTemplateUpdated       = "template.updated"
	TypeChannelUpdated        = "channel.updated"
)

// CanonicalEvent is the provider-neutral envelope published to NATS.
type CanonicalEvent struct {
	EventID    string            `json:"event_id"`
	Provider   string            `json:"provider"`
	TeamID     string            `json:"team_id,omitempty"`
	ChannelID  string            `json:"channel_id,omitempty"`
	Type       string            `json:"type"`
	OccurredAt time.Time         `json:"occurred_at"`
	Payload    json.RawMessage   `json:"payload"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// MetaSubject returns the NATS subject for a canonical Meta event type.
func MetaSubject(eventType string) string {
	switch eventType {
	case TypeMessageReceived:
		return "meta.message.received"
	case TypeMessageStatusSent:
		return "meta.message.status.sent"
	case TypeMessageStatusDelivered:
		return "meta.message.status.delivered"
	case TypeMessageStatusRead:
		return "meta.message.status.read"
	case TypeMessageStatusFailed:
		return "meta.message.status.failed"
	case TypeTemplateUpdated:
		return "meta.template.updated"
	case TypeChannelUpdated:
		return "meta.channel.updated"
	default:
		return "meta.unknown"
	}
}

// MetaDLQSubject returns dead-letter subject for a failure category.
func MetaDLQSubject(category string) string {
	return "meta.dlq." + category
}

// MetaConsumerDLQSubject returns backend consumer dead-letter subject.
func MetaConsumerDLQSubject(category string) string {
	return "meta.consumer.dlq." + category
}

// ConsumerDLQCategory maps canonical event type to consumer DLQ category.
func ConsumerDLQCategory(eventType string) string {
	switch eventType {
	case TypeMessageReceived:
		return "message"
	case TypeMessageStatusSent, TypeMessageStatusDelivered,
		TypeMessageStatusRead, TypeMessageStatusFailed:
		return "status"
	case TypeTemplateUpdated:
		return "template"
	case TypeChannelUpdated:
		return "channel"
	default:
		return "unknown"
	}
}
