package events

import "encoding/json"

// WhatsApp inbound message (normalized subset).
type WhatsAppMessageReceivedPayload struct {
	PhoneNumberID      string `json:"phone_number_id"`
	From               string `json:"from"`
	WAMID              string `json:"wamid"`
	Content            string `json:"content"`
	MessageType        string `json:"message_type"`
	ContactName        string `json:"contact_name,omitempty"`
	ContactWAID        string `json:"contact_wa_id,omitempty"`
	DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
	WABAID             string `json:"waba_id,omitempty"`
}

// WhatsApp delivery/read status update.
type WhatsAppMessageStatusPayload struct {
	PhoneNumberID string `json:"phone_number_id"`
	WAMID         string `json:"wamid"`
	Status        string `json:"status"`
	RecipientID   string `json:"recipient_id"`
	Timestamp     string `json:"timestamp"`
}

// WhatsApp template approval/status change.
type WhatsAppTemplateStatusPayload struct {
	WABAID         string `json:"waba_id"`
	TemplateID     string `json:"template_id"`
	TemplateName   string `json:"template_name"`
	TemplateLang   string `json:"template_language"`
	Event          string `json:"event"`
	Reason         string `json:"reason,omitempty"`
}

// WhatsApp account / coexistence sync payloads (history, contacts, echoes).
type WhatsAppAccountUpdatePayload struct {
	PhoneNumberID string `json:"phone_number_id"`
	Field         string `json:"field"`
	Raw           []byte `json:"raw"`
}

// Telegram message update.
type TelegramMessageReceivedPayload struct {
	UpdateID int64           `json:"update_id"`
	Message  json.RawMessage `json:"message"`
}

// Telegram callback query.
type TelegramCallbackReceivedPayload struct {
	UpdateID      int64           `json:"update_id"`
	CallbackQuery json.RawMessage `json:"callback_query"`
}
