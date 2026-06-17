package normalizer

import (
	"encoding/json"
	"fmt"

	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

type MetaEvent struct {
	Type    string
	Payload any
}

type metaWebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Field string          `json:"field"`
			Value json.RawMessage `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// NormalizeMeta splits a Meta WhatsApp webhook body into internal event types.
func NormalizeMeta(body []byte) ([]MetaEvent, error) {
	var payload metaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse meta webhook: %w", err)
	}

	var out []MetaEvent
	for _, entry := range payload.Entry {
		wabaID := entry.ID
		for _, change := range entry.Changes {
			switch change.Field {
			case "messages":
				items, err := parseInboundMessages(change.Value)
				if err != nil {
					return nil, err
				}
				for i := range items {
					items[i].WABAID = wabaID
					out = append(out, MetaEvent{Type: sharedevents.WhatsAppMessageReceived, Payload: items[i]})
				}
				statusItems, err := parseMessageStatuses(change.Value)
				if err != nil {
					return nil, err
				}
				for _, item := range statusItems {
					out = append(out, MetaEvent{Type: sharedevents.WhatsAppMessageStatus, Payload: item})
				}
			case "message_status", "statuses":
				items, err := parseMessageStatuses(change.Value)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					out = append(out, MetaEvent{Type: sharedevents.WhatsAppMessageStatus, Payload: item})
				}
			case "message_template_status_update":
				items, err := parseTemplateStatus(change.Value, wabaID)
				if err != nil {
					return nil, err
				}
				for _, item := range items {
					out = append(out, MetaEvent{Type: sharedevents.WhatsAppTemplateStatus, Payload: item})
				}
			case "account_update", "phone_number_name_update", "phone_number_quality_update",
				"smb_app_state_sync", "history", "smb_message_echoes":
				out = append(out, MetaEvent{
					Type: sharedevents.WhatsAppAccountUpdate,
					Payload: sharedevents.WhatsAppAccountUpdatePayload{
						PhoneNumberID: extractPhoneNumberID(change.Value),
						Field:         change.Field,
						Raw:           append(json.RawMessage(nil), change.Value...),
					},
				})
			default:
				out = append(out, MetaEvent{
					Type: sharedevents.WhatsAppAccountUpdate,
					Payload: sharedevents.WhatsAppAccountUpdatePayload{
						PhoneNumberID: extractPhoneNumberID(change.Value),
						Field:         change.Field,
						Raw:           append(json.RawMessage(nil), change.Value...),
					},
				})
			}
		}
	}
	return out, nil
}

func extractPhoneNumberID(raw json.RawMessage) string {
	var base struct {
		Metadata struct {
			PhoneNumberID string `json:"phone_number_id"`
		} `json:"metadata"`
	}
	_ = json.Unmarshal(raw, &base)
	return base.Metadata.PhoneNumberID
}

func parseInboundMessages(raw json.RawMessage) ([]sharedevents.WhatsAppMessageReceivedPayload, error) {
	var value struct {
		Metadata struct {
			DisplayPhoneNumber string `json:"display_phone_number"`
			PhoneNumberID      string `json:"phone_number_id"`
		} `json:"metadata"`
		Contacts []struct {
			Profile struct {
				Name string `json:"name"`
			} `json:"profile"`
			WAID string `json:"wa_id"`
		} `json:"contacts"`
		Messages []struct {
			From      string `json:"from"`
			ID        string `json:"id"`
			Type      string `json:"type"`
			Text      struct {
				Body string `json:"body"`
			} `json:"text"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}

	contactName := ""
	contactWAID := ""
	if len(value.Contacts) > 0 {
		contactName = value.Contacts[0].Profile.Name
		contactWAID = value.Contacts[0].WAID
	}

	items := make([]sharedevents.WhatsAppMessageReceivedPayload, 0, len(value.Messages))
	for _, msg := range value.Messages {
		content := msg.Text.Body
		if content == "" && msg.Type != "text" {
			content = fmt.Sprintf("[%s message]", msg.Type)
		}
		if content == "" {
			continue
		}
		name := contactName
		waID := contactWAID
		if waID == "" {
			waID = msg.From
		}
		items = append(items, sharedevents.WhatsAppMessageReceivedPayload{
			PhoneNumberID:      value.Metadata.PhoneNumberID,
			DisplayPhoneNumber: value.Metadata.DisplayPhoneNumber,
			From:               msg.From,
			WAMID:              msg.ID,
			Content:            content,
			MessageType:        msg.Type,
			ContactName:        name,
			ContactWAID:        waID,
		})
	}
	return items, nil
}

func parseMessageStatuses(raw json.RawMessage) ([]sharedevents.WhatsAppMessageStatusPayload, error) {
	var value struct {
		Metadata struct {
			PhoneNumberID string `json:"phone_number_id"`
		} `json:"metadata"`
		Statuses []struct {
			ID          string `json:"id"`
			Status      string `json:"status"`
			Timestamp   string `json:"timestamp"`
			RecipientID string `json:"recipient_id"`
		} `json:"statuses"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	items := make([]sharedevents.WhatsAppMessageStatusPayload, 0, len(value.Statuses))
	for _, st := range value.Statuses {
		items = append(items, sharedevents.WhatsAppMessageStatusPayload{
			PhoneNumberID: value.Metadata.PhoneNumberID,
			WAMID:         st.ID,
			Status:        st.Status,
			RecipientID:   st.RecipientID,
			Timestamp:     st.Timestamp,
		})
	}
	return items, nil
}

func parseTemplateStatus(raw json.RawMessage, wabaID string) ([]sharedevents.WhatsAppTemplateStatusPayload, error) {
	var value struct {
		MessageTemplateID   string `json:"message_template_id"`
		MessageTemplateName string `json:"message_template_name"`
		MessageTemplateLang string `json:"message_template_language"`
		Event               string `json:"event"`
		Reason              string `json:"reason"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return []sharedevents.WhatsAppTemplateStatusPayload{{
		WABAID:       wabaID,
		TemplateID:   value.MessageTemplateID,
		TemplateName: value.MessageTemplateName,
		TemplateLang: value.MessageTemplateLang,
		Event:        value.Event,
		Reason:       value.Reason,
	}}, nil
}
