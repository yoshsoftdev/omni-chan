# Phase 11.3 — Webhook Ingestion Service Review

**Date:** 2026-06-22  
**Verdict:** APPROVE WITH FINDINGS

## Summary

Dedicated `webhook-service` receives Meta webhooks, validates signatures, normalizes to canonical events, deduplicates via Redis, publishes to NATS JetStream (`meta.*` subjects), and ACKs Meta without PostgreSQL dependency.

## Event Model

**PASS**

- Canonical `CanonicalEvent` envelope in `shared/events/canonical.go`
- Meta normalizer at `internal/webhook/meta/normalizer/`
- Raw payload preserved; wamid, phone_number_id, template_id extracted

## Security

**PASS**

- GET challenge verification
- `X-Hub-Signature-256` validation before processing
- No token/message content in logs

## NATS Design

**PASS**

- Subjects: `meta.message.received`, `meta.message.status.*`, `meta.template.updated`, `meta.channel.updated`
- DLQ: `meta.dlq.*`
- Stream includes `meta.>` and legacy `events.>`

## Idempotency

**PASS**

- Redis SET NX with 24h TTL
- In-memory fallback for dev
- Duplicate webhooks suppressed

## Observability

**PASS**

- Metrics at `GET /api/v1/metrics`
- Structured JSON logs with event_id, event_type, phone_number_id
- Counters: requests, validation failures, published, publish failures, duplicates, DLQ

## Build & Test

| Check | Result |
|-------|--------|
| `go test ./...` (webhook-service) | PASS |
| `go test ./...` (shared) | PASS |
| No PostgreSQL in webhook path | PASS |

## Findings

1. `team_id` / `channel_id` empty at ingestion — consumers resolve in Phase 11.4 (by design)
2. Legacy `events.whatsapp.*` backend consumer not yet wired to `meta.*` subjects
3. Telegram webhook handler removed from stub; re-add in future phase if needed
4. Signature validation disabled when `META_APP_SECRET` unset (dev only)

## Recommendation

Proceed to **Phase 11.4 — Event Consumers** to subscribe to `meta.*` subjects and implement inbox/message/report updates.
