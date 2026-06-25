# WAMID Correlation

Campaign delivery tracking correlates Meta WhatsApp message IDs (`wamid`) to `campaign_recipients` rows.

## Lookup path

```
Meta status webhook
  → canonical event (payload.wamid)
  → FindRecipientByWAMIDForTeam(team_id, wamid)
  → campaign_recipients row
```

## Properties

| Requirement | Implementation |
|-------------|----------------|
| O(1) lookup | Unique index `idx_campaign_recipients_wamid` on `wamid` |
| Tenant safe | Query filters `team_id` from channel resolver |
| Set on send | Worker sets `wamid` when Meta returns message id (Phase 12) |

## Index

Migration `000012_broadcast_execution.up.sql`:

```sql
CREATE UNIQUE INDEX idx_campaign_recipients_wamid
  ON campaign_recipients (wamid, created_at)
  WHERE wamid IS NOT NULL;
```

## Dual path

The same `wamid` may exist in:

- `messages` (inbox) — updated by `message.Service.ApplyDeliveryStatus`
- `campaign_recipients` (broadcast) — updated by `delivery.Tracker`

Handlers run both paths; each no-ops when the row is not found.

## Idempotency

Consumer idempotency key: `provider:wamid:status`.

Recipient updates use `WHERE status = $from` so duplicate or regressive events do not change counters.
