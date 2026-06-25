# Meta Webhook Event Catalog

Events ingested by `webhook-service` and normalized to canonical types.

## Messages (`field: messages`)

| Meta type | Canonical type | Notes |
|-----------|----------------|-------|
| `text` | `message.received` | Body in raw payload |
| `image` | `message.received` | Media metadata preserved in raw |
| `video` | `message.received` | |
| `audio` | `message.received` | |
| `document` | `message.received` | |
| `sticker` | `message.received` | |
| `location` | `message.received` | |
| `contacts` | `message.received` | |

## Status (`statuses` in messages change)

| Meta status | Canonical type |
|-------------|----------------|
| `sent` | `message.status.sent` |
| `delivered` | `message.status.delivered` |
| `read` | `message.status.read` |
| `failed` | `message.status.failed` |

## Template

| Meta field | Canonical type |
|------------|----------------|
| `message_template_status_update` | `template.updated` |

## Channel / Account

| Meta field | Canonical type |
|------------|----------------|
| `phone_number_quality_update` | `channel.updated` |
| `account_update` | `channel.updated` |
| `phone_number_name_update` | `channel.updated` |
| `smb_app_state_sync` | `channel.updated` |
| `history` | `channel.updated` |
| `smb_message_echoes` | `channel.updated` |

Unknown fields are published as `channel.updated` with raw payload preserved.
