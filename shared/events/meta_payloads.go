package events

import "encoding/json"

// MetaMessagePayload is the canonical payload for message.received events.
type MetaMessagePayload struct {
	WABAID        string          `json:"waba_id"`
	PhoneNumberID string          `json:"phone_number_id"`
	WAMID         string          `json:"wamid"`
	ReplyToWAMID  string          `json:"reply_to_wamid,omitempty"`
	MessageType   string          `json:"message_type"`
	From          string          `json:"from"`
	ContactWAID   string          `json:"contact_wa_id,omitempty"`
	ContactName   string          `json:"contact_name,omitempty"`
	Timestamp     string          `json:"timestamp,omitempty"`
	DisplayPhone  string          `json:"display_phone_number,omitempty"`
	Raw           json.RawMessage `json:"raw"`
}

// MetaStatusPayload is the canonical payload for message.status.* events.
type MetaStatusPayload struct {
	PhoneNumberID string          `json:"phone_number_id"`
	WAMID         string          `json:"wamid"`
	Status        string          `json:"status"`
	RecipientID   string          `json:"recipient_id,omitempty"`
	Timestamp     string          `json:"timestamp,omitempty"`
	Raw           json.RawMessage `json:"raw"`
}

// MetaTemplatePayload is the canonical payload for template.updated events.
type MetaTemplatePayload struct {
	WABAID       string          `json:"waba_id"`
	TemplateID   string          `json:"template_id"`
	TemplateName string          `json:"template_name"`
	Language     string          `json:"language"`
	Event        string          `json:"event"`
	Reason       string          `json:"reason,omitempty"`
	Raw          json.RawMessage `json:"raw"`
}

// MetaChannelPayload is the canonical payload for channel.updated events.
type MetaChannelPayload struct {
	Field         string          `json:"field"`
	WABAID        string          `json:"waba_id"`
	PhoneNumberID string          `json:"phone_number_id"`
	Raw           json.RawMessage `json:"raw"`
}
