# Dead Letter Strategy

Failures during webhook ingestion are routed to DLQ NATS subjects without blocking Meta ACK.

## DLQ Triggers

| Failure | Category | Subject |
|---------|----------|---------|
| Malformed JSON / normalize error | `message` | `meta.dlq.message` |
| NATS publish failure | per event type | `meta.dlq.{category}` |
| Idempotency store error | per event type | `meta.dlq.{category}` |

## DLQ Message Format

```json
{
  "category": "status",
  "error": "publish failed: ...",
  "raw_body": { "...": "original webhook" },
  "event": { "...": "canonical event if available" },
  "at": "2026-06-22T10:00:00Z"
}
```

## Meta ACK Policy

`POST /webhooks/meta` returns `200 OK` after validation even if normalization/publish fails internally. Meta must not retry indefinitely; failures are recovered from DLQ.

## Consumer Phase

Phase 11.4 consumers do not process DLQ. Operational replay tooling is a future concern.
