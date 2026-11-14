package events

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
