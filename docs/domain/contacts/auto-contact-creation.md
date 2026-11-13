# Auto Contact Creation

Inbound Meta messages may arrive from unknown phone numbers.

## Flow

```
message.received
        ↓
extract from / contact_wa_id
        ↓
contacts.FindByPhone(team_id, phone)
        ↓
create if missing
```

## Defaults

| Field | Value |
|-------|-------|
| `source` | `meta_inbound` |
| `marketing_status` | `unknown` |
| `name` | contact profile name, else phone |

## Phone Normalization

Phones are normalized to E.164 (`+` prefix) for lookup and storage consistency.

## Idempotency

Concurrent events for the same phone rely on database uniqueness and duplicate-create fallback to `FindByPhone`.

## Scope

Auto-creation applies to inbound message consumers only. No quota enforcement at ingest (matches onboarding coexistence behavior).
