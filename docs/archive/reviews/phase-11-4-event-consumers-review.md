# Phase 11.4 — Event Consumers Review

**Date:** 2026-06-22  
**Verdict:** APPROVE WITH FINDINGS

## Summary

Backend consumers subscribe to `meta.*` JetStream subjects, resolve tenants, and persist conversations, messages, template updates, and channel health metadata.

## Tenant Resolution

**PASS**

- `FindByPhoneNumberID` / `FindByWABAID` on channel repository
- Unknown channels → `meta.consumer.dlq.channel`

## Conversation Creation

**PASS**

- Uniqueness: `(team_id, contact_id, channel_id)`
- Reuses existing conversation via `FindByContactChannel`

## Message Processing

**PASS**

- Inbound ingestion with `raw_payload` JSONB
- Idempotent by `wamid`
- Auto contact creation with `meta_inbound` source

## Status Tracking

**PASS**

- Monotonic progression: sent → delivered → read
- Duplicate status events suppressed via consumer idempotency + service guard

## Idempotency

**PASS**

- Redis SET NX (fallback: in-memory)
- Keys: message/wamid, status/wamid/status, template/id/ts, channel/field/phone/ts

## DLQ Strategy

**PASS**

- `meta.consumer.dlq.{message,status,template,channel}`
- Terminal failures ack after DLQ publish
- Transient repository errors NAK for JetStream retry (max 5)

## Observability

**PASS**

- Consumer metrics struct (`consumer_events_total`, etc.)
- Structured logs with event_id, team_id, channel_id — no message body

## Build & Test

| Check | Result |
|-------|--------|
| `go build ./...` | PASS |
| `go test ./internal/eventconsumer/...` | PASS |
| Legacy `events.>` consumer | PASS (parallel subscriber) |

## Findings

1. Unread counter / realtime inbox updates deferred to Phase 11.5
2. Campaign recipient status not updated (by design — later phase)
3. Migration `000018` required for `raw_payload`, `health_metadata`, `meta_inbound` source

## Recommendation

Proceed to **Phase 11.5 — Inbox Synchronization** for assignment rules, unread automation, team inbox views, and WebSocket realtime.
