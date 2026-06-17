package normalizer

import (
	"encoding/json"
	"fmt"

	sharedevents "github.com/omnichannel/omni-chan-v2/shared/events"
)

type TelegramEvent struct {
	Type    string
	Payload any
}

// NormalizeTelegram splits a Telegram webhook update.
func NormalizeTelegram(body []byte) ([]TelegramEvent, error) {
	var update struct {
		UpdateID      int64           `json:"update_id"`
		Message       json.RawMessage `json:"message"`
		CallbackQuery json.RawMessage `json:"callback_query"`
	}
	if err := json.Unmarshal(body, &update); err != nil {
		return nil, fmt.Errorf("parse telegram update: %w", err)
	}

	var out []TelegramEvent
	if len(update.Message) > 0 {
		out = append(out, TelegramEvent{
			Type: sharedevents.TelegramMessageReceived,
			Payload: sharedevents.TelegramMessageReceivedPayload{
				UpdateID: update.UpdateID,
				Message:  update.Message,
			},
		})
	}
	if len(update.CallbackQuery) > 0 {
		out = append(out, TelegramEvent{
			Type: sharedevents.TelegramCallbackReceived,
			Payload: sharedevents.TelegramCallbackReceivedPayload{
				UpdateID:      update.UpdateID,
				CallbackQuery: update.CallbackQuery,
			},
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("telegram update has no supported payload")
	}
	return out, nil
}
