# Event Envelope

Provider-neutral canonical event published to NATS by `webhook-service`.

## Schema

```json
{
  "event_id": "uuid",
  "provider": "meta",
  "team_id": "",
  "channel_id": "",
  "type": "message.received",
  "occurred_at": "2026-06-22T10:00:00Z",
  "payload": {},
  "metadata": {
    "waba_id": "...",
    "phone_number_id": "...",
    "wamid": "..."
  }
}
```

## Field Rules

| Field | Rule |
|-------|------|
| `event_id` | UUID v4, unique per published event |
| `provider` | `meta`, `telegram`, `line`, etc. |
| `team_id` | Empty at ingestion — resolved by consumers (Phase 11.4) |
| `channel_id` | Empty at ingestion — resolved via `phone_number_id` lookup in consumer |
| `type` | Provider-neutral dot notation |
| `occurred_at` | Meta timestamp when available, else receive time |
| `payload` | Provider-specific JSON; always includes `raw` where applicable |
| `metadata` | Routing hints only — no secrets |

## Multi-Provider Design

Meta-specific structures stay inside `payload`. Envelope shape is identical for future Instagram, Telegram, Line integrations.

## Legacy Envelope

Backend consumers may still use `shared/events.Envelope` (`events.whatsapp.*`) until Phase 11.4 migration. New ingestion publishes `meta.*` subjects with `CanonicalEvent`.
