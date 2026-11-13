# Idempotency Strategy

Meta may deliver the same webhook multiple times. `webhook-service` suppresses duplicate NATS publishes.

## Storage

- **Production:** Redis `SET NX` with TTL (default 24h)
- **Fallback:** In-memory map when Redis unavailable (dev only)

## Keys

| Event | Key pattern |
|-------|-------------|
| Message | `meta:message:{wamid}` |
| Status | `meta:status:{wamid}:{status}` |
| Template | `meta:template:{template_id}:{occurred_at_unix}` |
| Channel | `meta:channel:{field}:{phone_number_id}:{occurred_at_unix}` |

Redis key prefix: `wh:idemp:`

## Behavior

1. Normalize webhook → N canonical events
2. For each event, compute idempotency key
3. `SET NX` — if key exists, skip publish (duplicate)
4. If new, publish to NATS

## TTL

Configurable via `IDEMPOTENCY_TTL` (default `24h`). Matches Meta redelivery window.

## No PostgreSQL

Idempotency does not use PostgreSQL — webhook path remains DB-free.
